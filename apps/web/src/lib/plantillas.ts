/**
 * Las 16 plantillas de tienda del handoff (design/templates-data.js): colores, tipografía base,
 * layout y la tienda de ejemplo. Generado una vez desde ese archivo; si cambia el handoff, regenerar.
 */
export type Plantilla = {
	id: string;
	nombre: string;
	rubro: string;
	layout: 'ed' | 'grid' | 'cat';
	tienda: string;
	bg: string;
	tx: string;
	ac: string;
	serif: boolean;
	radio: string;
	lema: string;
	oscura: boolean;
};

export const PLANTILLAS: Plantilla[] = [
	{
		id: 'ceibo',
		nombre: 'Ceibo',
		rubro: 'Moda',
		layout: 'ed',
		tienda: 'Tienda Lumbre',
		bg: '#F4EFEA',
		tx: '#1C1A17',
		ac: '#1C1A17',
		serif: true,
		radio: '2px',
		lema: 'Nueva temporada',
		oscura: false
	},
	{
		id: 'mostaza',
		nombre: 'Mostaza',
		rubro: 'Gastronomía',
		layout: 'grid',
		tienda: 'Café Bruma',
		bg: '#FBF3E4',
		tx: '#3B2416',
		ac: '#C9821A',
		serif: false,
		radio: '14px',
		lema: 'Tostado esta semana',
		oscura: false
	},
	{
		id: 'yerbal',
		nombre: 'Yerbal',
		rubro: 'Almacén',
		layout: 'cat',
		tienda: 'Almacén Verde',
		bg: '#EEF1E8',
		tx: '#1E2B22',
		ac: '#2F5D50',
		serif: false,
		radio: '10px',
		lema: 'Orgánico y a granel',
		oscura: false
	},
	{
		id: 'pampa',
		nombre: 'Pampa',
		rubro: 'Moda',
		layout: 'grid',
		tienda: 'PMPA Store',
		bg: '#141414',
		tx: '#FFFFFF',
		ac: '#E6FF4D',
		serif: false,
		radio: '6px',
		lema: 'Drop 07 · ya disponible',
		oscura: true
	},
	{
		id: 'rosedal',
		nombre: 'Rosedal',
		rubro: 'Belleza',
		layout: 'ed',
		tienda: 'Rosa Mosqueta',
		bg: '#F8EDEB',
		tx: '#3A1F25',
		ac: '#C26A7A',
		serif: true,
		radio: '18px',
		lema: 'Rutina de otoño',
		oscura: false
	},
	{
		id: 'quebracho',
		nombre: 'Quebracho',
		rubro: 'Hogar',
		layout: 'ed',
		tienda: 'Casa Nido',
		bg: '#EFE9E1',
		tx: '#2B2520',
		ac: '#8A5A3E',
		serif: true,
		radio: '4px',
		lema: 'Muebles hechos a mano',
		oscura: false
	},
	{
		id: 'andes',
		nombre: 'Andes',
		rubro: 'Tecnología',
		layout: 'grid',
		tienda: 'Sonido Sur',
		bg: '#F2F4F7',
		tx: '#0F1720',
		ac: '#2563EB',
		serif: false,
		radio: '16px',
		lema: 'Envío en 24 h',
		oscura: false
	},
	{
		id: 'malbec',
		nombre: 'Malbec',
		rubro: 'Bebidas',
		layout: 'ed',
		tienda: 'Vinoteca Tres',
		bg: '#1E0E12',
		tx: '#F4E9E1',
		ac: '#C24A5A',
		serif: true,
		radio: '2px',
		lema: 'Selección del mes',
		oscura: true
	},
	{
		id: 'patio',
		nombre: 'Patio',
		rubro: 'Mascotas',
		layout: 'grid',
		tienda: 'Huellitas',
		bg: '#FFF7E8',
		tx: '#2A1E10',
		ac: '#F08A24',
		serif: false,
		radio: '22px',
		lema: 'Todo para tu compañero',
		oscura: false
	},
	{
		id: 'miga',
		nombre: 'Miga',
		rubro: 'Gastronomía',
		layout: 'cat',
		tienda: 'Panadería Lola',
		bg: '#FBF1E6',
		tx: '#3A2414',
		ac: '#B5651D',
		serif: true,
		radio: '12px',
		lema: 'Masa madre todos los días',
		oscura: false
	},
	{
		id: 'taller',
		nombre: 'Taller',
		rubro: 'Artesanías',
		layout: 'ed',
		tienda: 'Madera Viva',
		bg: '#F3EEE6',
		tx: '#2A231C',
		ac: '#6B4A2F',
		serif: false,
		radio: '8px',
		lema: 'Piezas únicas',
		oscura: false
	},
	{
		id: 'hora',
		nombre: 'Hora',
		rubro: 'Accesorios',
		layout: 'grid',
		tienda: 'Relojería Norte',
		bg: '#F5F5F2',
		tx: '#161616',
		ac: '#B08D57',
		serif: true,
		radio: '0px',
		lema: 'Clásicos que duran',
		oscura: false
	},
	{
		id: 'litoral',
		nombre: 'Litoral',
		rubro: 'Regionales',
		layout: 'cat',
		tienda: 'Mate & Co',
		bg: '#F3E9DC',
		tx: '#3B2416',
		ac: '#6B3A22',
		serif: false,
		radio: '14px',
		lema: 'El mate de cada mañana',
		oscura: false
	},
	{
		id: 'torno',
		nombre: 'Torno',
		rubro: 'Arte y diseño',
		layout: 'ed',
		tienda: 'Estudio Arcilla',
		bg: '#EDE6DD',
		tx: '#2C2621',
		ac: '#A4623D',
		serif: true,
		radio: '24px',
		lema: 'Cerámica de autor',
		oscura: false
	},
	{
		id: 'tijera',
		nombre: 'Tijera',
		rubro: 'Servicios',
		layout: 'grid',
		tienda: 'Navaja',
		bg: '#F6F4F1',
		tx: '#1E1B2E',
		ac: '#5B3FD6',
		serif: false,
		radio: '18px',
		lema: 'Reservá tu turno online',
		oscura: false
	},
	{
		id: 'lente',
		nombre: 'Lente',
		rubro: 'Accesorios',
		layout: 'grid',
		tienda: 'Óptica Faro',
		bg: '#FFF8EC',
		tx: '#1F1A12',
		ac: '#E0A31B',
		serif: false,
		radio: '12px',
		lema: '2×1 en anteojos de sol',
		oscura: false
	}
];

export const plantilla = (id: string) => PLANTILLAS.find((p) => p.id === id);
