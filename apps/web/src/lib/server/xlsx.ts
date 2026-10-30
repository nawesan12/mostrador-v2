import { zip } from './zip.ts';

/**
 * Planilla .xlsx mínima (una hoja, encabezado en negrita) armada a mano sobre el ZIP de
 * `zip.ts`: alcanza para "Excel" en Mis datos sin sumar una dependencia. Los textos van como
 * `inlineStr`, así que no hace falta la tabla de textos compartidos.
 */
type Celda = string | number | null;

const xml = (s: string) =>
	s
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		// Caracteres de control que XML no admite.
		// eslint-disable-next-line no-control-regex
		.replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f]/g, '');

/** A, B, …, Z, AA, AB… */
export function columna(i: number) {
	let s = '';
	for (let n = i + 1; n > 0; n = Math.floor((n - 1) / 26))
		s = String.fromCharCode(65 + ((n - 1) % 26)) + s;
	return s;
}

function hoja(filas: Celda[][]) {
	const cuerpo = filas
		.map((fila, r) => {
			const celdas = fila
				.map((v, c) => {
					const ref = `${columna(c)}${r + 1}`;
					const estilo = r === 0 ? ' s="1"' : '';
					if (v === null || v === '') return '';
					if (typeof v === 'number' && Number.isFinite(v))
						return `<c r="${ref}"${estilo}><v>${v}</v></c>`;
					return `<c r="${ref}"${estilo} t="inlineStr"><is><t xml:space="preserve">${xml(String(v))}</t></is></c>`;
				})
				.join('');
			return `<row r="${r + 1}">${celdas}</row>`;
		})
		.join('');
	return (
		'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
		'<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">' +
		(filas.length > 1
			? '<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>'
			: '') +
		`<sheetData>${cuerpo}</sheetData></worksheet>`
	);
}

/** El .xlsx con las filas (la primera es el encabezado). `nombre` es el de la pestaña. */
export function xlsx(filas: Celda[][], nombre = 'Hoja 1'): Uint8Array {
	const enc = new TextEncoder();
	const pestana = xml(nombre.replace(/[[\]:*?/\\]/g, ' ').slice(0, 31) || 'Hoja 1');
	const archivos: [string, string][] = [
		[
			'[Content_Types].xml',
			'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
				'<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">' +
				'<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>' +
				'<Default Extension="xml" ContentType="application/xml"/>' +
				'<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>' +
				'<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>' +
				'<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>' +
				'</Types>'
		],
		[
			'_rels/.rels',
			'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
				'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
				'<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>' +
				'</Relationships>'
		],
		[
			'xl/workbook.xml',
			'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
				'<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">' +
				`<sheets><sheet name="${pestana}" sheetId="1" r:id="rId1"/></sheets></workbook>`
		],
		[
			'xl/_rels/workbook.xml.rels',
			'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
				'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
				'<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>' +
				'<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>' +
				'</Relationships>'
		],
		[
			'xl/styles.xml',
			'<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
				'<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">' +
				'<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>' +
				'<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills>' +
				'<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>' +
				'<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>' +
				'<cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/></cellXfs>' +
				'<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>' +
				'</styleSheet>'
		],
		['xl/worksheets/sheet1.xml', hoja(filas)]
	];
	return zip(archivos.map(([n, t]) => ({ nombre: n, datos: enc.encode(t) })));
}

export const TIPO_XLSX = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
