// Package visitas guarda y lee las visitas de la tienda online, medidas por Tu Local Web sin
// cookies ni datos personales (decisiones §36).
//
// La web cuenta cada página vista de /t/<slug> en su servidor y manda lotes a
// POST /v1/interno/visitas con la clave del BFF. Cada vista trae un `visitante`: un hash de IP,
// navegador, tienda y una sal del día que sólo vive en la memoria de la web. Acá se guarda el
// agregado por tienda, día (hora de Argentina) y página; el hash sólo sirve para contar personas
// distintas en el día y se borra al día siguiente (app.purgar_visitas).
package visitas

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5"
)

// Argentina no tiene horario de verano: UTC−3 fijo.
var argentina = time.FixedZone("ART", -3*60*60)

// MaxLote: vistas por pedido. La web manda de a menos; lo que pase se descarta.
const MaxLote = 2000

// maxTiendasPorLote: tiendas distintas que se procesan por lote (una transacción cada una).
const maxTiendasPorLote = 200

// Paginas que se miden. 'tienda' es el total del día y no viene de la web.
var Paginas = map[string]bool{"inicio": true, "catalogo": true, "producto": true, "carrito": true, "checkout": true}

// Vista es una página vista, como la manda la web.
type Vista struct {
	// Tienda: el slug.
	Tienda string `json:"tienda"`
	Pagina string `json:"pagina"`
	// Producto: el slug del producto (sólo en la página de producto).
	Producto string `json:"producto,omitempty"`
	// Visitante: 32 caracteres hexadecimales (16 bytes de hash).
	Visitante string `json:"visitante"`
	// Duena: el id de la tienda de la sesión del panel de quien mira, si tiene una. Si es la misma
	// tienda, la vista no se cuenta (la dueña o su equipo mirando su propia tienda).
	Duena string `json:"duena,omitempty"`
}

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Dia: la fecha de Argentina de `t`, a las 0 h UTC (como la devuelve Postgres para un date).
func Dia(t time.Time) time.Time {
	a := t.In(argentina)
	return time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
}

type clave struct{ pagina, producto string }

// Registrar suma un lote de vistas al día de `ahora`. Devuelve cuántas contó (las de tiendas o
// productos que no existen, las mal formadas y las de la propia dueña no cuentan).
func Registrar(ctx context.Context, pool tx, vs []Vista, ahora time.Time) (int, error) {
	if len(vs) > MaxLote {
		vs = vs[:MaxLote]
	}
	porTienda := map[string][]Vista{}
	var orden []string
	for _, v := range vs {
		if !Paginas[v.Pagina] || len(v.Visitante) != 32 || len(v.Tienda) > 60 || len(v.Producto) > 60 {
			continue
		}
		if (v.Pagina == "producto") != (v.Producto != "") {
			continue
		}
		if _, ok := porTienda[v.Tienda]; !ok {
			if len(orden) == maxTiendasPorLote {
				continue
			}
			orden = append(orden, v.Tienda)
		}
		porTienda[v.Tienda] = append(porTienda[v.Tienda], v)
	}
	dia := Dia(ahora)
	total := 0
	for _, slug := range orden {
		n, err := registrarEnTienda(ctx, pool, slug, porTienda[slug], dia)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func registrarEnTienda(ctx context.Context, pool tx, slug string, vs []Vista, dia time.Time) (int, error) {
	var id *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return 0, err
	}
	if id == nil {
		// Una tienda que no existe, cerrada o pausada (o la vitrina de una plantilla): no se mide.
		return 0, nil
	}
	contadas := 0
	err := pool.EnTienda(ctx, *id, func(t pgx.Tx) error {
		// Los productos por slug: sólo cuentan los que existen (nadie llena la tabla con inventos).
		var slugs []string
		for _, v := range vs {
			if v.Producto != "" {
				slugs = append(slugs, v.Producto)
			}
		}
		productos := map[string]string{}
		if len(slugs) > 0 {
			filas, err := t.Query(ctx, `SELECT slug, id::text FROM productos WHERE tienda_id = $1 AND slug = ANY($2)`, *id, slugs)
			if err != nil {
				return err
			}
			for filas.Next() {
				var s, pid string
				if err := filas.Scan(&s, &pid); err != nil {
					filas.Close()
					return err
				}
				productos[s] = pid
			}
			filas.Close()
			if err := filas.Err(); err != nil {
				return err
			}
		}

		vistas := map[clave]int{}
		var orden []clave
		suma := func(k clave) {
			if _, ok := vistas[k]; !ok {
				orden = append(orden, k)
			}
			vistas[k]++
		}
		// Únicos: cada visitante una vez por página y por día (y una vez en el total de la tienda).
		type visto struct {
			k clave
			h string
		}
		yaVisto := map[visto]bool{}
		var vClaves []string
		var vHashes [][]byte
		for _, v := range vs {
			if v.Duena != "" && v.Duena == *id {
				continue
			}
			h, err := hex.DecodeString(v.Visitante)
			if err != nil || len(h) != 16 {
				continue
			}
			k := clave{pagina: v.Pagina}
			if v.Pagina == "producto" {
				pid, ok := productos[v.Producto]
				if !ok {
					continue
				}
				k.producto = pid
			}
			contadas++
			suma(clave{pagina: "tienda"})
			suma(k)
			for _, x := range []clave{{pagina: "tienda"}, k} {
				if yaVisto[visto{x, v.Visitante}] {
					continue
				}
				yaVisto[visto{x, v.Visitante}] = true
				vClaves = append(vClaves, textoClave(x))
				vHashes = append(vHashes, h)
			}
		}
		if contadas == 0 {
			return nil
		}
		nuevos := map[string]int{}
		filas, err := t.Query(ctx, `INSERT INTO visitas_vistos (tienda_id, dia, clave, visitante)
			SELECT $1::uuid, $2::date, c, v FROM unnest($3::text[], $4::bytea[]) AS u(c, v)
			ON CONFLICT DO NOTHING RETURNING clave`, *id, dia, vClaves, vHashes)
		if err != nil {
			return err
		}
		for filas.Next() {
			var c string
			if err := filas.Scan(&c); err != nil {
				filas.Close()
				return err
			}
			nuevos[c]++
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}

		paginas := make([]string, len(orden))
		prods := make([]string, len(orden))
		cuantas := make([]int32, len(orden))
		unicos := make([]int32, len(orden))
		for i, k := range orden {
			paginas[i], prods[i] = k.pagina, k.producto
			cuantas[i], unicos[i] = int32(vistas[k]), int32(nuevos[textoClave(k)])
		}
		_, err = t.Exec(ctx, `INSERT INTO visitas (tienda_id, dia, pagina, producto_id, vistas, visitantes)
			SELECT $1::uuid, $2::date, p, nullif(pr, '')::uuid, n, u FROM unnest($3::text[], $4::text[], $5::int[], $6::int[]) AS x(p, pr, n, u)
			ON CONFLICT (tienda_id, dia, pagina, producto_id) DO UPDATE SET
				vistas = visitas.vistas + EXCLUDED.vistas, visitantes = visitas.visitantes + EXCLUDED.visitantes`,
			*id, dia, paginas, prods, cuantas, unicos)
		return err
	})
	if err != nil {
		return 0, err
	}
	return contadas, nil
}

func textoClave(k clave) string {
	if k.producto != "" {
		return "producto:" + k.producto
	}
	return k.pagina
}

// Purgar borra los vistos de días anteriores y las visitas de hace más de 400 días (todas las
// tiendas). La corre el worker una vez por hora.
func Purgar(ctx context.Context, pool interface {
	EnTx(context.Context, func(pgx.Tx) error) error
}) (int64, error) {
	var n int64
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.purgar_visitas()`).Scan(&n)
	})
	return n, err
}
