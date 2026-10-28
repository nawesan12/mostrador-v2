package partners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/partners"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func usuario(t *testing.T, b *pruebadb.Bases, email, nombre string) string {
	t.Helper()
	var id string
	err := b.App.EnTx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `INSERT INTO auth.usuarios (email, nombre) VALUES ($1, $2) RETURNING id`, email, nombre).Scan(&id)
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// tienda crea una tienda con su origen y, como superusuario, le pone plan y antigüedad.
func tienda(t *testing.T, b *pruebadb.Bases, admin *pgx.Conn, slug string, o tiendas.Origen, plan string, dias int) string {
	t.Helper()
	ctx := context.Background()
	id := pruebadb.CrearTienda(t, b.App, slug)
	if o != (tiendas.Origen{}) {
		if ok, err := tiendas.FijarOrigen(ctx, b.App, id, o); err != nil || !ok {
			t.Fatalf("origen de %s: %v %v", slug, ok, err)
		}
	}
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET plan = $2, creada_en = now() - make_interval(days => $3) WHERE id = $1`, id, plan, dias); err != nil {
		t.Fatal(err)
	}
	return id
}

func conectar(t *testing.T, b *pruebadb.Bases) *pgx.Conn {
	t.Helper()
	admin, err := pgx.Connect(context.Background(), b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close(context.Background()) })
	return admin
}

func TestAtribucionPorRefYCupon(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin := conectar(t, b)
	ana := usuario(t, b, "ana@pampacode.com.ar", "Ana")
	codigo, err := partners.Aprobar(ctx, b.App, partners.Aprobacion{Email: "ana@pampacode.com.ar", Nombre: "Pampa Code"})
	if err != nil || codigo != "pampa-code" {
		t.Fatalf("aprobar: %q %v", codigo, err)
	}
	// Por el ref (sin distinguir mayúsculas), por el cupón, de otro partner y sin origen.
	tienda(t, b, admin, "bruma", tiendas.Origen{Ref: "Pampa-Code"}, "pro", 40)
	tienda(t, b, admin, "nido", tiendas.Origen{Cupon: "PAMPACODE20"}, "emprende", 30)
	tienda(t, b, admin, "ajena", tiendas.Origen{Ref: "otro-estudio"}, "pro", 40)
	tienda(t, b, admin, "sola", tiendas.Origen{}, "pro", 40)
	tienda(t, b, admin, "flyer", tiendas.Origen{Fuente: "flyer"}, "pro", 40)

	p, err := partners.Leer(ctx, b.App, ana, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tiendas) != 2 {
		t.Fatalf("tiendas: %+v", p.Tiendas)
	}
	nombres := map[string]bool{}
	for _, x := range p.Tiendas {
		nombres[x.Slug] = true
		if x.Estado != partners.EstadoActiva {
			t.Errorf("%s: %s", x.Slug, x.Estado)
		}
	}
	if !nombres["bruma"] || !nombres["nido"] {
		t.Fatalf("%v", nombres)
	}
	if p.Partner.Cupon != "PAMPACODE20" || p.Partner.Nivel != "partner" || p.Partner.Porcentaje != 20 {
		t.Fatalf("%+v", p.Partner)
	}

	// Clics: un ref de verdad suma (en mayúsculas también); uno inventado no.
	for range 3 {
		if ok, err := partners.SumarClic(ctx, b.App, "PAMPA-CODE"); err != nil || !ok {
			t.Fatalf("clic: %v %v", ok, err)
		}
	}
	if ok, _ := partners.SumarClic(ctx, b.App, "nadie"); ok {
		t.Fatal("un ref inventado sumó")
	}
	p, _ = partners.Leer(ctx, b.App, ana, time.Now())
	if len(p.Clics) != 1 || p.Clics[0].N != 3 {
		t.Fatalf("clics: %+v", p.Clics)
	}

	// Una cuenta que no es partner no ve tiendas ni ficha.
	beto := usuario(t, b, "beto@mail.com", "Beto")
	if _, err := partners.Leer(ctx, b.App, beto, time.Now()); !errors.Is(err, partners.ErrNoEsPartner) {
		t.Fatalf("no partner: %v", err)
	}
}

func TestComisionDelMesYNivelPorTiendas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin := conectar(t, b)
	ana := usuario(t, b, "ana@pampacode.com.ar", "Ana")
	if _, err := partners.Aprobar(ctx, b.App, partners.Aprobacion{Email: "ana@pampacode.com.ar", Nombre: "Pampa Code"}); err != nil {
		t.Fatal(err)
	}
	ref := tiendas.Origen{Ref: "pampa-code"}
	// Cinco Pro activas (una con más de un año: ya no genera), una en prueba, una gratis y una de baja.
	for i, dias := range []int{30, 60, 90, 120, 400} {
		tienda(t, b, admin, "pro-"+string(rune('a'+i)), ref, "pro", dias)
	}
	tienda(t, b, admin, "prueba", ref, "pro", 3)
	tienda(t, b, admin, "gratis", ref, "gratis", 60)
	baja := tienda(t, b, admin, "baja", ref, "pro", 60)
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET cancelada_en = now() WHERE id = $1`, baja); err != nil {
		t.Fatal(err)
	}

	p, err := partners.Leer(ctx, b.App, ana, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	// Cinco activas: Plata (25 %) aunque el equipo la aprobó como Partner.
	if p.Partner.NivelEfectivo != "plata" || p.Partner.Porcentaje != 25 {
		t.Fatalf("nivel: %+v", p.Partner)
	}
	var total int64
	estados := map[string]int{}
	for _, x := range p.Tiendas {
		total += x.ComisionMes
		estados[x.Estado]++
	}
	if total != 4*12_475 {
		t.Fatalf("comisión del mes: %d, quería %d (la de 400 días ya no cuenta)", total, 4*12_475)
	}
	if estados["activa"] != 5 || estados["prueba"] != 1 || estados["gratis"] != 1 || estados["baja"] != 1 {
		t.Fatalf("estados: %v", estados)
	}

	// La liquidación congela el detalle; la factura se sube hasta que se paga.
	mes := time.Now().Format("2006-01")
	g, err := partners.Liquidar(ctx, b.App, "ana@pampacode.com.ar", mes, time.Now())
	if err != nil || g.Monto != 4*12_475 || len(g.Detalle) != 4 {
		t.Fatalf("liquidar: %+v %v", g, err)
	}
	if err := partners.SubirFactura(ctx, b.App, ana, mes, "http://inseguro.com/f.pdf"); err == nil {
		t.Fatal("factura sin https")
	}
	if err := partners.SubirFactura(ctx, b.App, ana, mes, "https://drive.google.com/f.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := partners.CambiarEstado(ctx, b.App, "ana@pampacode.com.ar", mes, "pagada"); err != nil {
		t.Fatal(err)
	}
	if err := partners.SubirFactura(ctx, b.App, ana, mes, "https://drive.google.com/otra.pdf"); !errors.Is(err, partners.ErrSinPago) {
		t.Fatalf("factura de un pago pagado: %v", err)
	}
	if _, err := partners.Liquidar(ctx, b.App, "ana@pampacode.com.ar", mes, time.Now()); err == nil {
		t.Fatal("se rehizo un mes pagado")
	}
	p, _ = partners.Leer(ctx, b.App, ana, time.Now())
	if len(p.Pagos) != 1 || p.Pagos[0].Estado != "pagada" || p.Pagos[0].PagadoEn == nil || *p.Pagos[0].FacturaURL != "https://drive.google.com/f.pdf" {
		t.Fatalf("pagos: %+v", p.Pagos)
	}
}

func TestDosPartnersNoSeCruzan(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin := conectar(t, b)
	ana := usuario(t, b, "ana@pampacode.com.ar", "Ana")
	beto := usuario(t, b, "beto@brote.com", "Beto")
	for _, a := range []partners.Aprobacion{{Email: "ana@pampacode.com.ar", Nombre: "Pampa Code"}, {Email: "beto@brote.com", Nombre: "Estudio Brote", Nivel: "oro"}} {
		if _, err := partners.Aprobar(ctx, b.App, a); err != nil {
			t.Fatal(err)
		}
	}
	tienda(t, b, admin, "de-ana", tiendas.Origen{Ref: "pampa-code"}, "pro", 40)
	tienda(t, b, admin, "de-beto", tiendas.Origen{Ref: "estudio-brote"}, "pro", 40)
	mes := time.Now().Format("2006-01")
	for _, e := range []string{"ana@pampacode.com.ar", "beto@brote.com"} {
		if _, err := partners.Liquidar(ctx, b.App, e, mes, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := partners.GuardarCobro(ctx, b.App, beto, partners.Cobro{Tipo: "alias", Valor: "brote.mp", Titular: "Beto", Cuit: "20123456786", CondicionIva: "monotributo"}); err != nil {
		t.Fatal(err)
	}

	t.Run("cada uno ve lo suyo", func(t *testing.T) {
		pa, _ := partners.Leer(ctx, b.App, ana, time.Now())
		pb, _ := partners.Leer(ctx, b.App, beto, time.Now())
		if len(pa.Tiendas) != 1 || pa.Tiendas[0].Slug != "de-ana" || len(pb.Tiendas) != 1 || pb.Tiendas[0].Slug != "de-beto" {
			t.Fatalf("ana %+v beto %+v", pa.Tiendas, pb.Tiendas)
		}
		if pa.Partner.Cobro != nil || pb.Partner.Cobro == nil {
			t.Fatal("el cobro de Beto se ve en Ana")
		}
		if pb.Partner.Porcentaje != 30 || pb.Tiendas[0].ComisionMes != 14_970 {
			t.Fatalf("beto oro: %+v", pb)
		}
	})

	t.Run("con la persona fijada, RLS no deja ver ni tocar lo del otro", func(t *testing.T) {
		var fichas, pagos int
		var tocadas int64
		err := b.App.EnUsuario(ctx, ana, func(tx pgx.Tx) error {
			if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM partners.partners), (SELECT count(*) FROM partners.pagos)`).Scan(&fichas, &pagos); err != nil {
				return err
			}
			tag, err := tx.Exec(ctx, `UPDATE partners.pagos SET factura_url = 'https://x.com' WHERE partner_id <> (SELECT id FROM partners.partners)`)
			tocadas = tag.RowsAffected()
			return err
		})
		if err != nil || fichas != 1 || pagos != 1 || tocadas != 0 {
			t.Fatalf("err=%v fichas=%d pagos=%d tocadas=%d", err, fichas, pagos, tocadas)
		}
	})

	t.Run("sin persona no se ve nada", func(t *testing.T) {
		var n int
		err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM partners.partners) + (SELECT count(*) FROM partners.pagos)
				+ (SELECT count(*) FROM partners.clics) + (SELECT count(*) FROM partners.tiendas_referidas())`).Scan(&n)
		})
		if err != nil || n != 0 {
			t.Fatalf("err=%v filas=%d", err, n)
		}
	})

	t.Run("la función no deja la puerta abierta", func(t *testing.T) {
		// Después de listar las referidas, la misma transacción no ve tiendas ajenas.
		var n int
		err := b.App.EnUsuario(ctx, ana, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `SELECT count(*) FROM partners.tiendas_referidas()`); err != nil {
				return err
			}
			return tx.QueryRow(ctx, `SELECT count(*) FROM tiendas`).Scan(&n)
		})
		if err != nil || n != 0 {
			t.Fatalf("err=%v tiendas=%d", err, n)
		}
	})

	t.Run("staff lee pero no escribe", func(t *testing.T) {
		var n int
		if err := b.Staff.QueryRow(ctx, `SELECT count(*) FROM partners.pagos`).Scan(&n); err != nil || n != 2 {
			t.Fatalf("staff: %v %d", err, n)
		}
		if _, err := b.Staff.Exec(ctx, `DELETE FROM partners.pagos`); err == nil {
			t.Fatal("staff borró pagos")
		}
	})
}

func TestSembrarLaDemo(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin := conectar(t, b)
	ana := usuario(t, b, "ana@pampacode.com.ar", "Ana")
	if _, err := partners.Sembrar(ctx, admin, "ana@pampacode.com.ar", time.Now()); err != nil {
		t.Fatal(err)
	}
	p, err := partners.Leer(ctx, b.App, ana, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tiendas) != 19 || len(p.Pagos) != 3 || p.Partner.NivelEfectivo != "plata" || len(p.Clics) != 3 {
		t.Fatalf("tiendas=%d pagos=%d nivel=%s clics=%d", len(p.Tiendas), len(p.Pagos), p.Partner.NivelEfectivo, len(p.Clics))
	}
}
