package partners

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
)

// Tiendas de la demo de partners: nombre, plan, ciclo, hace cuántos días se dio de alta y si se dio
// de baja. Las mismas que la demo del sandbox (apps/web/src/lib/server/partners/demo.ts).
var tiendasDemo = []struct {
	nombre, plan, ciclo string
	dias                int
	baja                bool
}{
	{"Almacén Verde", "pro", "mes", 320, false},
	{"Café Bruma", "pro", "mes", 290, false},
	{"Rosa Mosqueta", "emprende", "mes", 260, false},
	{"Casa Nido", "pro", "anio", 240, false},
	{"Sonido Sur", "negocio", "mes", 210, false},
	{"Vinoteca Tres", "pro", "mes", 180, false},
	{"Huellitas", "emprende", "mes", 150, false},
	{"Madera Viva", "pro", "mes", 130, false},
	{"Relojería Norte", "emprende", "anio", 110, false},
	{"Mate & Co", "pro", "mes", 90, false},
	{"Estudio Arcilla", "pro", "mes", 70, false},
	{"Óptica Faro", "emprende", "mes", 50, false},
	{"Navaja Barbería", "pro", "mes", 40, false},
	{"Panadería Lola", "emprende", "mes", 30, false},
	{"Tienda Lumbre", "gratis", "mes", 6, false},
	{"PMPA Store", "gratis", "mes", 3, false},
	{"Kiosco Central", "gratis", "mes", 11, false},
	{"Librería Faro", "gratis", "mes", 400, false},
	{"Dulce Taller", "emprende", "mes", 200, true},
}

// Sembrar arma la demo del panel de partners para una cuenta que ya existe: la aprueba como
// Plata ("Pampa Code"), le da 19 tiendas referidas con distintos planes y antigüedades, clics y
// los pagos de los últimos tres meses. Corre como superusuario (mostradorctl partner-demo), sólo
// en local: escribe fechas de alta en el pasado, que la app nunca puede tocar.
func Sembrar(ctx context.Context, admin *pgx.Conn, email string, ahora time.Time) (string, error) {
	tx, err := admin.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var usuario string
	if err := tx.QueryRow(ctx, `SELECT id FROM auth.usuarios WHERE email = $1`, email).Scan(&usuario); err != nil {
		return "", fmt.Errorf("la cuenta %s no existe (creala en /ingresar): %w", email, err)
	}
	perfil, _ := json.Marshal(Perfil{
		Nombre: "Pampa Code", Ciudad: "Rosario",
		Descripcion:    "Armamos tiendas para marcas del Litoral: integraciones, migraciones y fotos de producto.",
		Especialidades: []string{"desarrollo", "diseno", "fotografia"},
		Casos:          []Caso{{"Café Bruma", "https://mostrador.iamnaw.dev/t/cafe-bruma"}, {"Casa Nido", "https://mostrador.iamnaw.dev/t/casa-nido"}},
		Email:          email, Whatsapp: "5493415550123", Web: "https://pampacode.com.ar",
	})
	var partner string
	if err := tx.QueryRow(ctx, `INSERT INTO partners.partners (usuario_id, nivel, codigo, cupon, perfil)
		VALUES ($1, 'plata', 'pampa-code', 'PAMPACODE20', $2)
		ON CONFLICT (usuario_id) DO UPDATE SET nivel = 'plata', perfil = excluded.perfil RETURNING id`, usuario, perfil).Scan(&partner); err != nil {
		return "", err
	}
	for i, t := range tiendasDemo {
		slug := tiendas.Slug(t.nombre) + "-demo"
		origen, _ := json.Marshal(map[string]string{"ref": "pampa-code", "fuente": "partner"})
		if i%4 == 3 {
			origen, _ = json.Marshal(map[string]string{"cupon": "PAMPACODE20"})
		}
		alta := ahora.Add(-time.Duration(t.dias) * 24 * time.Hour)
		var baja *time.Time
		if t.baja {
			b := ahora.Add(-20 * 24 * time.Hour)
			baja = &b
		}
		if _, err := tx.Exec(ctx, `INSERT INTO app.slugs (slug) VALUES ($1) ON CONFLICT DO NOTHING`, slug); err != nil {
			return "", err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tiendas (slug, nombre, rubro, plantilla, plan, plan_ciclo, creada_en, cancelada_en, origen)
			VALUES ($1, $2, 'regionales', 'litoral', $3, $4, $5, $6, $7)
			ON CONFLICT (slug) DO UPDATE SET plan = excluded.plan, plan_ciclo = excluded.plan_ciclo,
				creada_en = excluded.creada_en, cancelada_en = excluded.cancelada_en, origen = excluded.origen`,
			slug, t.nombre, t.plan, t.ciclo, alta, baja, origen); err != nil {
			return "", err
		}
	}
	mes := primeroDeMes(ahora)
	// Los pagos de los tres meses anteriores: dos pagados (con factura) y el último aprobado.
	var fs []fila
	for _, t := range tiendasDemo {
		fs = append(fs, fila{id: tiendas.Slug(t.nombre), nombre: t.nombre, plan: t.plan, ciclo: t.ciclo,
			creada: ahora.Add(-time.Duration(t.dias) * 24 * time.Hour), baja: t.baja})
	}
	for i, estado := range []string{"aprobada", "pagada", "pagada"} {
		m := mes.AddDate(0, -(i + 1), 0)
		_, ts := calcular("plata", fs, ahora, m)
		detalle := []Linea{}
		var monto int64
		for _, t := range ts {
			if t.ComisionMes > 0 {
				detalle = append(detalle, Linea{TiendaID: t.ID, Tienda: t.Nombre, Plan: t.Plan, Comision: t.ComisionMes})
				monto += t.ComisionMes
			}
		}
		d, _ := json.Marshal(detalle)
		var factura *string
		if estado == "pagada" {
			f := fmt.Sprintf("https://drive.google.com/file/d/factura-pampa-code-%s", m.Format("2006-01"))
			factura = &f
		}
		if _, err := tx.Exec(ctx, `INSERT INTO partners.pagos (partner_id, mes, estado, monto, detalle, factura_url, aprobado_en, pagado_en)
			VALUES ($1, $2, $3, $4, $5, $6, now(), CASE WHEN $3 = 'pagada' THEN $7::timestamptz END)
			ON CONFLICT (partner_id, mes) DO UPDATE SET estado = excluded.estado, monto = excluded.monto,
				detalle = excluded.detalle, factura_url = excluded.factura_url, pagado_en = excluded.pagado_en`,
			partner, m, estado, monto, d, factura, FechaDePago(m)); err != nil {
			return "", err
		}
	}
	for i, n := range []int{212, 187, 164} {
		if _, err := tx.Exec(ctx, `INSERT INTO partners.clics (partner_id, mes, n) VALUES ($1, $2, $3)
			ON CONFLICT (partner_id, mes) DO UPDATE SET n = excluded.n`, partner, mes.AddDate(0, -i, 0), n); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return partner, nil
}
