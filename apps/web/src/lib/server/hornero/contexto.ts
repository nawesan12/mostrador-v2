/**
 * Lo que Tincho sabe para responder: la tienda, el tablero de hoy y de la semana, ventas por día,
 * productos, clientes y canales. Va en el mensaje de sistema como JSON compacto (no inventa nada que
 * no esté acá).
 */
import type { ContextoHornero, Tablero, Tienda } from '../datos/tipos.ts';

const TZ = 'America/Argentina/Buenos_Aires';

export type DatosHornero = {
	tienda: Pick<Tienda, 'nombre' | 'rubro'>;
	hola: string | null;
	hoy: Tablero;
	semana: Tablero;
	contexto: ContextoHornero;
};

export function datosParaElModelo(d: DatosHornero) {
	const { hoy, semana, contexto } = d;
	// Con la API real las visitas no se miden (conversión 0) y no hay WhatsApp para mandar mensajes:
	// al modelo no se le pasa un 0 % ni un costo por mensaje que después repetiría como dato.
	const midoVisitas = semana.kpis.conversion.actual > 0;
	return {
		// Hoy hasta ahora y ayer completo; el detalle de cada día va en ventasPorDia.
		hoyHastaAhora: { ventas: hoy.ventas.actual, pedidos: hoy.kpis.pedidos.actual },
		ayerDiaCompleto: contexto.ventasPorDia.at(-2) ?? null,
		ultimos7Dias: {
			ventas: semana.ventas.actual,
			semanaAnterior: semana.ventas.anterior,
			pedidos: semana.kpis.pedidos.actual,
			ticketPromedio: semana.kpis.ticket.actual,
			conversionPorcentaje: midoVisitas ? semana.kpis.conversion.actual : 'no se mide todavía',
			clientesNuevos: semana.kpis.clientesNuevos.actual,
			masVendidos: semana.masVendidos.map((p) => `${p.nombre}: ${p.unidades} u`)
		},
		metaDelMes: semana.meta,
		ventasPorDia: contexto.ventasPorDia,
		productos: contexto.productos,
		clientes: contexto.clientes,
		ventasPorCanal30Dias: contexto.canales30d,
		pendientes: semana.paraHacer,
		carritosAbandonados: { cantidad: semana.hornero.carritos, monto: semana.hornero.carritosMonto },
		...(midoVisitas ? { costoMensajeWhatsApp: semana.hornero.costoMensaje } : {}),
		resenas: { promedio: semana.resenas.promedio, nuevas: semana.resenas.nuevas }
	};
}

export function sistema(d: DatosHornero, ahora = new Date()) {
	const cuando = new Intl.DateTimeFormat('es-AR', {
		timeZone: TZ,
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		hour: '2-digit',
		minute: '2-digit'
	}).format(ahora);
	return [
		`Sos Tincho, el asistente de la tienda «${d.tienda.nombre}» (${d.tienda.rubro}) en Tu Local Web, una plataforma argentina de venta online.`,
		d.hola ? `Hablás con ${d.hola}, la dueña.` : 'Hablás con quien maneja la tienda.',
		`Ahora es ${cuando} (hora de Argentina).`,
		'Hablá en español rioplatense, cálido y concreto, como alguien del equipo. Respondé en 1 a 4 oraciones (máximo 90 palabras), sin títulos ni listas largas.',
		'Marcá con **negrita** sólo el dato clave. Escribí la plata así: $ 12.345.',
		'Usá sólo los DATOS de abajo. Si algo no está, decí que todavía no lo tenés; nunca inventes números, nombres ni fechas.',
		'Si te piden hacer algo (subir precios, mandar una campaña, pausar productos), contá cómo quedaría con los datos y en qué sección del panel se hace. Desde el chat no podés hacerlo ni dejarlo preparado: nunca digas que lo hiciste ni que quedó listo.',
		'Si te preguntan algo que no tiene que ver con la tienda, respondé con amabilidad que sólo ayudás con la tienda.',
		`DATOS: ${JSON.stringify(datosParaElModelo(d))}`
	].join('\n');
}
