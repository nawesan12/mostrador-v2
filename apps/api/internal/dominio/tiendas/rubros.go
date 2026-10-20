// Package tiendas: alta y lectura de la tienda, y los rubros.
package tiendas

// Rubro con su plantilla y tienda de ejemplo. Tiene que coincidir con packages/contrato/rubros.json
// (lo verifica TestRubrosIgualAlContrato).
type Rubro struct {
	ID        string `json:"id"`
	Nombre    string `json:"nombre"`
	Plantilla string `json:"plantilla"`
	Tienda    string `json:"tienda"`
}

// Rubros en el orden del contrato.
var Rubros = []Rubro{
	{"moda", "Moda", "ceibo", "Tienda Lumbre"},
	{"gastronomia", "Gastronomía", "mostaza", "Café Bruma"},
	{"almacen", "Almacén", "yerbal", "Almacén Verde"},
	{"belleza", "Belleza", "rosedal", "Rosa Mosqueta"},
	{"hogar", "Hogar", "quebracho", "Casa Nido"},
	{"tecnologia", "Tecnología", "andes", "Sonido Sur"},
	{"bebidas", "Bebidas", "malbec", "Vinoteca Tres"},
	{"mascotas", "Mascotas", "patio", "Huellitas"},
	{"artesanias", "Artesanías", "taller", "Madera Viva"},
	{"accesorios", "Accesorios", "hora", "Relojería Norte"},
	{"regionales", "Regionales", "litoral", "Mate & Co"},
	{"arte", "Arte y diseño", "torno", "Estudio Arcilla"},
	{"servicios", "Servicios", "tijera", "Navaja"},
}

// BuscarRubro devuelve el rubro por id.
func BuscarRubro(id string) (Rubro, bool) {
	for _, r := range Rubros {
		if r.ID == id {
			return r, true
		}
	}
	return Rubro{}, false
}
