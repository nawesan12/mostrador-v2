package equipo

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
)

// TiendaBreve (components/schemas/TiendaBreve).
type TiendaBreve struct {
	ID     string  `json:"id"`
	Nombre string  `json:"nombre"`
	Slug   string  `json:"slug"`
	Fecha  string  `json:"fecha"`
	Motivo *string `json:"motivo"`
}

// Alerta (components/schemas/AlertaEquipo).
type Alerta struct {
	Tipo     string  `json:"tipo"`
	Titulo   string  `json:"titulo"`
	Detalle  string  `json:"detalle"`
	TiendaID *string `json:"tiendaId"`
	Consulta *string `json:"consulta"`
}

// Inicio (components/schemas/InicioEquipo).
type Inicio struct {
	Kpis struct {
		TiendasActivas        int   `json:"tiendasActivas"`
		NuevasSemana          int   `json:"nuevasSemana"`
		PruebasPorVencer      int   `json:"pruebasPorVencer"`
		Mrr                   int64 `json:"mrr"`
		BajasMes              int   `json:"bajasMes"`
		ConsultasSinResponder int   `json:"consultasSinResponder"`
		PostulacionesNuevas   int   `json:"postulacionesNuevas"`
	} `json:"kpis"`
	Pruebas []TiendaBreve `json:"pruebas"`
	Bajas   []TiendaBreve `json:"bajas"`
	Alertas []Alerta      `json:"alertas"`
}

func inicioDeMes(ahora time.Time) time.Time {
	a := ahora.In(ar)
	return time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, ar)
}

// Inicio arma la pantalla de inicio del backoffice.
func (s *Servicio) Inicio(ctx context.Context) (Inicio, error) {
	ahora := s.ahora()
	in := Inicio{Pruebas: []TiendaBreve{}, Bajas: []TiendaBreve{}, Alertas: []Alerta{}}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		ts, err := leerTiendas(ctx, tx, "")
		if err != nil {
			return err
		}
		mes := inicioDeMes(ahora)
		for _, t := range ts {
			e := t.estado(ahora)
			if t.Cerrada == nil && t.Pausada == nil {
				in.Kpis.TiendasActivas++
			}
			if ahora.Sub(t.Creada) < 7*24*time.Hour {
				in.Kpis.NuevasSemana++
			}
			in.Kpis.Mrr += t.mrr(ahora)
			if e == "prueba" && t.PruebaHasta.Sub(ahora) <= 72*time.Hour {
				in.Kpis.PruebasPorVencer++
				in.Pruebas = append(in.Pruebas, TiendaBreve{ID: t.ID, Nombre: t.Nombre, Slug: t.Slug, Fecha: fecha(t.PruebaHasta)})
			}
			baja := t.Cerrada
			if baja == nil {
				baja = t.Cancelada
			}
			if baja != nil && !baja.Before(mes) {
				in.Kpis.BajasMes++
				in.Bajas = append(in.Bajas, TiendaBreve{ID: t.ID, Nombre: t.Nombre, Slug: t.Slug, Fecha: fecha(*baja), Motivo: t.BajaMotivo})
			}
			if e == "cobro_rechazado" {
				id := t.ID
				in.Alertas = append(in.Alertas, Alerta{Tipo: "cobro_rechazado", Titulo: "Cobro rechazado: " + t.Nombre,
					Detalle: "Mercado Pago no pudo cobrar el abono " + nombresPlan[t.Plan] + ". Escribile a la titular.", TiendaID: &id})
			}
		}
		sort.Slice(in.Pruebas, func(i, j int) bool { return in.Pruebas[i].Fecha < in.Pruebas[j].Fecha })
		sort.Slice(in.Bajas, func(i, j int) bool { return in.Bajas[i].Fecha > in.Bajas[j].Fecha })
		in.Pruebas = in.Pruebas[:min(8, len(in.Pruebas))]
		in.Bajas = in.Bajas[:min(8, len(in.Bajas))]

		filas, err := tx.Query(ctx, `SELECT t.id::text, t.nombre FROM mp_conexiones c JOIN tiendas t ON t.id = c.tienda_id
			WHERE c.reautorizar AND t.cerrada_en IS NULL ORDER BY t.nombre LIMIT 10`)
		if err != nil {
			return err
		}
		mp, err := pgx.CollectRows(filas, pgx.RowToStructByPos[struct{ ID, Nombre string }])
		if err != nil {
			return err
		}
		for _, t := range mp {
			id := t.ID
			in.Alertas = append(in.Alertas, Alerta{Tipo: "integracion", Titulo: "Mercado Pago desconectado: " + t.Nombre,
				Detalle: "La conexión venció y la tienda no puede cobrar online hasta reconectarla.", TiendaID: &id})
		}
		if err := tx.QueryRow(ctx, `SELECT
			(SELECT count(*) FROM soporte.consultas WHERE estado = 'nueva') + (SELECT count(*) FROM soporte.arrepentimientos WHERE estado = 'recibido'),
			(SELECT count(*) FROM soporte.postulaciones_partner WHERE estado = 'nueva')`).
			Scan(&in.Kpis.ConsultasSinResponder, &in.Kpis.PostulacionesNuevas); err != nil {
			return err
		}
		filas, err = tx.Query(ctx, `SELECT codigo, nombre, creado_en FROM soporte.arrepentimientos
			WHERE estado IN ('recibido', 'en_curso') ORDER BY creado_en LIMIT 50`)
		if err != nil {
			return err
		}
		arr, err := pgx.CollectRows(filas, pgx.RowToStructByPos[struct {
			Codigo, Nombre string
			Creado         time.Time
		}])
		if err != nil {
			return err
		}
		hoy := diasHabiles(ahora, 0)
		for _, a := range arr {
			plazo := diasHabiles(a.Creado, 10)
			if !plazo.After(diasHabiles(hoy, 3)) {
				c := "arrepentimiento/" + a.Codigo
				in.Alertas = append(in.Alertas, Alerta{Tipo: "arrepentimiento", Titulo: "Devolución por vencer: " + a.Codigo,
					Detalle: "A " + a.Nombre + " hay que devolverle el total antes del " + plazo.Format("02/01") + ".", Consulta: &c})
			}
		}
		return nil
	})
	return in, err
}

// DefTiendas: el motor de listas para todas las tiendas.
var DefTiendas = listas.Def{
	Buscar: func(r listas.Registro) string {
		return r["nombre"].(string) + " " + r["slug"].(string) + " " + r["titular"].(string) + " " + r["email"].(string)
	},
	Segmento:   "estado",
	Facetas:    []string{"plan", "rubro", "plantilla", "partner"},
	Rango:      "dias",
	RangoMenor: true,
	Ordenes: []listas.Orden{
		{Clave: "recientes", Menor: func(a, b listas.Registro) bool {
			return a["creadaEn"].(time.Time).After(b["creadaEn"].(time.Time))
		}},
		{Clave: "nombre", Menor: func(a, b listas.Registro) bool { return a["nombre"].(string) < b["nombre"].(string) }},
		{Clave: "mrr", Menor: func(a, b listas.Registro) bool { return a["mrr"].(int64) > b["mrr"].(int64) }},
	},
	Resumen: func(todos, filtrados []listas.Registro) map[string]float64 {
		r := map[string]float64{"mrr": 0, "prueba": 0, "activas": 0}
		for _, x := range filtrados {
			r["mrr"] += float64(x["mrr"].(int64))
		}
		for _, x := range todos {
			switch x["estado"] {
			case "prueba":
				r["prueba"]++
			case "activa":
				r["activas"]++
			}
		}
		return r
	},
}

// Tiendas devuelve la lista de todas las tiendas con los filtros del motor de listas.
func (s *Servicio) Tiendas(ctx context.Context, f listas.Filtros) (listas.Lista, error) {
	ahora := s.ahora()
	var regs []listas.Registro
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		ts, err := leerTiendas(ctx, tx, "")
		for _, t := range ts {
			regs = append(regs, listas.Registro{
				"id": t.ID, "nombre": t.Nombre, "slug": t.Slug, "titular": t.Titular, "email": t.Email,
				"plan": t.Plan, "ciclo": t.Ciclo, "estado": t.estado(ahora), "rubro": t.Rubro, "plantilla": t.Plantilla,
				"partner": t.Partner, "creadaEn": t.Creada, "dias": int(ahora.Sub(t.Creada).Hours() / 24), "mrr": t.mrr(ahora),
			})
		}
		return err
	})
	if err != nil {
		return listas.Lista{}, err
	}
	return listas.Aplicar(regs, DefTiendas, f), nil
}

// Nota interna (components/schemas/NotaEquipo).
type Nota struct {
	ID       string    `json:"id"`
	Texto    string    `json:"texto"`
	Autor    string    `json:"autor"`
	CreadaEn time.Time `json:"creadaEn"`
}

// Ficha (components/schemas/FichaTiendaEquipo).
type Ficha struct {
	Tienda struct {
		ID         string    `json:"id"`
		Nombre     string    `json:"nombre"`
		Slug       string    `json:"slug"`
		Rubro      string    `json:"rubro"`
		Plantilla  string    `json:"plantilla"`
		CreadaEn   time.Time `json:"creadaEn"`
		Estado     string    `json:"estado"`
		Partner    string    `json:"partner"`
		MotivoBaja *string   `json:"motivoBaja"`
	} `json:"tienda"`
	Negocio map[string]string `json:"negocio"`
	Titular *struct {
		Nombre string    `json:"nombre"`
		Email  string    `json:"email"`
		Desde  time.Time `json:"desde"`
	} `json:"titular"`
	Personas int `json:"personas"`
	Plan     struct {
		Actual       string     `json:"actual"`
		Ciclo        string     `json:"ciclo"`
		Precio       int64      `json:"precio"`
		EnPrueba     bool       `json:"enPrueba"`
		PruebaHasta  time.Time  `json:"pruebaHasta"`
		ProximoCobro *time.Time `json:"proximoCobro"`
		Suscripcion  *string    `json:"suscripcion"`
		Credito      int64      `json:"credito"`
	} `json:"plan"`
	Uso struct {
		Productos  int `json:"productos"`
		PedidosMes int `json:"pedidosMes"`
		PacksMes   int `json:"packsMes"`
	} `json:"uso"`
	Integraciones []Integracion `json:"integraciones"`
	Dominios      []Dominio     `json:"dominios"`
	Actividad     []Actividad   `json:"actividad"`
	Notas         []Nota        `json:"notas"`
}

// Integracion conectada y su salud.
type Integracion struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Salud  string `json:"salud"`
}

// Dominio propio.
type Dominio struct {
	Nombre     string `json:"nombre"`
	Verificado bool   `json:"verificado"`
	Principal  bool   `json:"principal"`
}

// Actividad reciente de la tienda (su auditoría).
type Actividad struct {
	ID     string    `json:"id"`
	Cuando time.Time `json:"cuando"`
	Quien  string    `json:"quien"`
	Accion string    `json:"accion"`
}

// Ficha de una tienda; el acceso queda en la auditoría del equipo.
func (s *Servicio) Ficha(ctx context.Context, m Miembro, id string, cl cuentas.Cliente) (Ficha, error) {
	ahora := s.ahora()
	var f Ficha
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		ts, err := leerTiendas(ctx, tx, "WHERE t.id::text = $1", id)
		if err != nil {
			return err
		}
		if len(ts) == 0 {
			return ErrNoExiste
		}
		t := ts[0]
		f.Tienda.ID, f.Tienda.Nombre, f.Tienda.Slug, f.Tienda.Rubro, f.Tienda.Plantilla = t.ID, t.Nombre, t.Slug, t.Rubro, t.Plantilla
		f.Tienda.CreadaEn, f.Tienda.Estado, f.Tienda.Partner, f.Tienda.MotivoBaja = t.Creada, t.estado(ahora), t.Partner, t.BajaMotivo
		f.Plan.Actual, f.Plan.Ciclo, f.Plan.Precio = t.Plan, t.Ciclo, Mensual(t.Plan, t.Ciclo)
		f.Plan.EnPrueba, f.Plan.PruebaHasta = ahora.Before(t.PruebaHasta), t.PruebaHasta
		f.Plan.ProximoCobro, f.Plan.Suscripcion, f.Plan.Credito = t.proximoCobro(ahora), t.Suscripcion, t.Credito

		var negocio map[string]any
		var packs map[string]int
		var ga4, pixel *string
		mes := ahora.In(ar).Format("2006-01")
		if err := tx.QueryRow(ctx, `SELECT negocio, packs, ga4, pixel_meta,
				(SELECT count(*) FROM auth.membresias WHERE tienda_id = t.id),
				(SELECT count(*) FROM productos WHERE tienda_id = t.id),
				(SELECT count(*) FROM pedidos WHERE tienda_id = t.id AND creado_en >= $2)
			FROM tiendas t WHERE t.id::text = $1`, id, inicioDeMes(ahora)).
			Scan(&negocio, &packs, &ga4, &pixel, &f.Personas, &f.Uso.Productos, &f.Uso.PedidosMes); err != nil {
			return err
		}
		f.Uso.PacksMes = packs[mes]
		f.Negocio = map[string]string{}
		for k, v := range negocio {
			if str, ok := v.(string); ok && str != "" {
				f.Negocio[k] = str
			}
		}
		var tit struct {
			Nombre string    `json:"nombre"`
			Email  string    `json:"email"`
			Desde  time.Time `json:"desde"`
		}
		err = tx.QueryRow(ctx, `SELECT u.nombre, u.email::text, m.creada_en FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id
			WHERE m.tienda_id::text = $1 AND m.rol = 'duenio' ORDER BY m.creada_en LIMIT 1`, id).Scan(&tit.Nombre, &tit.Email, &tit.Desde)
		if err == nil {
			f.Titular = &tit
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		f.Integraciones = []Integracion{}
		salud := func(re bool) string {
			if re {
				return "reconectar"
			}
			return "ok"
		}
		var re bool
		if err := tx.QueryRow(ctx, `SELECT reautorizar FROM mp_conexiones WHERE tienda_id::text = $1`, id).Scan(&re); err == nil {
			f.Integraciones = append(f.Integraciones, Integracion{"mercadopago", "Mercado Pago", salud(re)})
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err := tx.QueryRow(ctx, `SELECT reautorizar FROM ml_conexiones WHERE tienda_id::text = $1`, id).Scan(&re); err == nil {
			f.Integraciones = append(f.Integraciones, Integracion{"mercadolibre", "Mercado Libre", salud(re)})
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if ga4 != nil {
			f.Integraciones = append(f.Integraciones, Integracion{"ga4", "Google Analytics", "ok"})
		}
		if pixel != nil {
			f.Integraciones = append(f.Integraciones, Integracion{"meta", "Píxel de Meta", "ok"})
		}

		filas, err := tx.Query(ctx, `SELECT nombre, verificado_en IS NOT NULL, principal FROM dominios WHERE tienda_id::text = $1 ORDER BY principal DESC, nombre`, id)
		if err != nil {
			return err
		}
		if f.Dominios, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Dominio]); err != nil {
			return err
		}
		filas, err = tx.Query(ctx, `SELECT id::text, creado_en, actor_tipo, accion FROM auditoria WHERE tienda_id::text = $1
			ORDER BY creado_en DESC, id DESC LIMIT 12`, id)
		if err != nil {
			return err
		}
		if f.Actividad, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Actividad]); err != nil {
			return err
		}
		filas, err = tx.Query(ctx, `SELECT n.id::text, n.texto, m.nombre, n.creada_en FROM equipo.notas n
			JOIN equipo.miembros m ON m.id = n.miembro_id WHERE n.tienda_id::text = $1 ORDER BY n.creada_en DESC LIMIT 50`, id)
		if err != nil {
			return err
		}
		if f.Notas, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Nota]); err != nil {
			return err
		}
		return anotar(ctx, tx, m, cl, "ver_ficha", "Vio la ficha de "+t.Nombre, t.ID, nil)
	})
	return f, err
}

// Notar agrega una nota interna a la tienda.
func (s *Servicio) Notar(ctx context.Context, m Miembro, id, texto string, cl cuentas.Cliente) (Nota, error) {
	n := Nota{Texto: texto, Autor: m.Nombre}
	if l := len([]rune(texto)); l < 1 || l > 1000 {
		return n, ErrInvalido{"texto", "La nota tiene que tener entre 1 y 1000 caracteres."}
	}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		var nombre string
		err := tx.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id::text = $1`, id).Scan(&nombre)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO equipo.notas (tienda_id, miembro_id, texto) VALUES ($1::uuid, $2, $3)
			RETURNING id::text, creada_en`, id, m.ID, texto).Scan(&n.ID, &n.CreadaEn); err != nil {
			return err
		}
		return anotar(ctx, tx, m, cl, "nota", "Dejó una nota en "+nombre, id, nil)
	})
	return n, err
}

// Suscripcion de una tienda con plan pago (FacturacionEquipo.suscripciones).
type Suscripcion struct {
	TiendaID     string     `json:"tiendaId"`
	Nombre       string     `json:"nombre"`
	Plan         string     `json:"plan"`
	Ciclo        string     `json:"ciclo"`
	Monto        int64      `json:"monto"`
	Estado       string     `json:"estado"`
	ProximoCobro *time.Time `json:"proximoCobro"`
	Credito      int64      `json:"credito"`
}

// Facturacion (components/schemas/FacturacionEquipo).
type Facturacion struct {
	Kpis struct {
		Mrr          int64 `json:"mrr"`
		CobrosMes    int64 `json:"cobrosMes"`
		Activas      int   `json:"activas"`
		ConProblemas int   `json:"conProblemas"`
		Creditos     int64 `json:"creditos"`
	} `json:"kpis"`
	Suscripciones []Suscripcion `json:"suscripciones"`
	// Comprobantes de ARCA: vacío hasta conectar ARCA.
	Comprobantes []Comprobante `json:"comprobantes"`
}

// Comprobante de Tu Local Web a una tienda (ARCA).
type Comprobante struct {
	Numero string `json:"numero"`
	Tienda string `json:"tienda"`
	Fecha  string `json:"fecha"`
	Monto  int64  `json:"monto"`
}

// Facturacion: los abonos de todas las tiendas.
func (s *Servicio) Facturacion(ctx context.Context) (Facturacion, error) {
	ahora := s.ahora()
	var f Facturacion
	f.Suscripciones = []Suscripcion{}
	f.Comprobantes = []Comprobante{}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		ts, err := leerTiendas(ctx, tx, "WHERE t.cerrada_en IS NULL ORDER BY t.nombre")
		if err != nil {
			return err
		}
		finMes := inicioDeMes(ahora).AddDate(0, 1, 0)
		for _, t := range ts {
			f.Kpis.Creditos += t.Credito
			if t.Plan == "gratis" {
				continue
			}
			e := t.estado(ahora)
			f.Kpis.Mrr += t.mrr(ahora)
			p := t.proximoCobro(ahora)
			switch e {
			case "activa":
				f.Kpis.Activas++
			case "cobro_rechazado":
				f.Kpis.ConProblemas++
			}
			if p != nil && (e == "activa" || e == "prueba") && p.Before(finMes) {
				f.Kpis.CobrosMes += Cobro(t.Plan, t.Ciclo)
			}
			f.Suscripciones = append(f.Suscripciones, Suscripcion{TiendaID: t.ID, Nombre: t.Nombre, Plan: t.Plan, Ciclo: t.Ciclo,
				Monto: Cobro(t.Plan, t.Ciclo), Estado: e, ProximoCobro: p, Credito: t.Credito})
		}
		return nil
	})
	return f, err
}

// Postulacion (components/schemas/PostulacionEquipo).
type Postulacion struct {
	ID             string    `json:"id"`
	Nombre         string    `json:"nombre"`
	Email          string    `json:"email"`
	Agencia        string    `json:"agencia"`
	Web            *string   `json:"web"`
	Especialidades []string  `json:"especialidades"`
	Estado         string    `json:"estado"`
	Motivo         *string   `json:"motivo"`
	CreadaEn       time.Time `json:"creadaEn"`
}

// PartnerActivo (components/schemas/PartnerActivoEquipo).
type PartnerActivo struct {
	Codigo   string `json:"codigo"`
	Nombre   string `json:"nombre"`
	Nivel    string `json:"nivel"`
	Tiendas  int    `json:"tiendas"`
	Activas  int    `json:"activas"`
	Comision int64  `json:"comision"`
}

// Partners: postulaciones y partners con tiendas referidas.
type Partners struct {
	Postulaciones []Postulacion   `json:"postulaciones"`
	Activos       []PartnerActivo `json:"activos"`
}

// NivelDe por tiendas activas: Plata desde 5, Oro desde 20. Comisión: 20, 25 y 30 %.
func NivelDe(activas int) (string, int64) {
	switch {
	case activas >= 20:
		return "oro", 30
	case activas >= 5:
		return "plata", 25
	}
	return "partner", 20
}

// Partners lee las postulaciones y arma los partners activos desde el ref con que llegó cada
// tienda (/empezar?ref=…). Cuando entre el paquete partners, el nombre y el nivel asignado salen de
// partners.partners.
func (s *Servicio) Partners(ctx context.Context) (Partners, error) {
	ahora := s.ahora()
	p := Partners{Postulaciones: []Postulacion{}, Activos: []PartnerActivo{}}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		filas, err := tx.Query(ctx, `SELECT id::text, nombre, email::text, agencia, web, especialidades, estado, motivo, creada_en
			FROM soporte.postulaciones_partner ORDER BY (estado = 'nueva') DESC, creada_en DESC LIMIT 500`)
		if err != nil {
			return err
		}
		if p.Postulaciones, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Postulacion]); err != nil {
			return err
		}
		ts, err := leerTiendas(ctx, tx, "WHERE t.origen ? 'ref'")
		if err != nil {
			return err
		}
		por := map[string]*PartnerActivo{}
		mrr := map[string]int64{}
		for _, t := range ts {
			if t.Partner == "" {
				continue
			}
			a := por[t.Partner]
			if a == nil {
				a = &PartnerActivo{Codigo: t.Partner, Nombre: t.Partner}
				por[t.Partner] = a
			}
			a.Tiendas++
			if m := t.mrr(ahora); m > 0 {
				a.Activas++
				mrr[t.Partner] += m
			}
		}
		for k, a := range por {
			var pct int64
			a.Nivel, pct = NivelDe(a.Activas)
			a.Comision = mrr[k] * pct / 100
			p.Activos = append(p.Activos, *a)
		}
		sort.Slice(p.Activos, func(i, j int) bool {
			if p.Activos[i].Comision != p.Activos[j].Comision {
				return p.Activos[i].Comision > p.Activos[j].Comision
			}
			return p.Activos[i].Codigo < p.Activos[j].Codigo
		})
		return nil
	})
	return p, err
}

// MiembroBreve (components/schemas/MiembroEquipo).
type MiembroBreve struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Rol    string `json:"rol"`
}

// Registro de la auditoría del equipo.
type Registro struct {
	ID       string       `json:"id"`
	Cuando   time.Time    `json:"cuando"`
	Miembro  MiembroBreve `json:"miembro"`
	Accion   string       `json:"accion"`
	Texto    string       `json:"texto"`
	TiendaID *string      `json:"tiendaId"`
	Tienda   *string      `json:"tienda"`
}

// Auditoria (components/schemas/AuditoriaEquipo).
type Auditoria struct {
	Registros []Registro     `json:"registros"`
	Miembros  []MiembroBreve `json:"miembros"`
}

func miembros(ctx context.Context, tx pgx.Tx, todos bool) ([]MiembroBreve, error) {
	filas, err := tx.Query(ctx, `SELECT id::text, nombre, rol FROM equipo.miembros WHERE activo OR $1 ORDER BY nombre`, todos)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, pgx.RowToStructByPos[MiembroBreve])
}

// Auditoria: los últimos 300 registros, con filtros.
func (s *Servicio) Auditoria(ctx context.Context, miembro, accion, tiendaID string) (Auditoria, error) {
	var a Auditoria
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		filas, err := tx.Query(ctx, `
			SELECT a.id::text, a.creado_en, m.id::text, m.nombre, m.rol, a.accion, a.texto, a.tienda_id::text, t.nombre
			FROM equipo.auditoria a JOIN equipo.miembros m ON m.id = a.miembro_id
			LEFT JOIN tiendas t ON t.id = a.tienda_id
			WHERE ($1 = '' OR a.miembro_id::text = $1) AND ($2 = '' OR a.accion = $2) AND ($3 = '' OR a.tienda_id::text = $3)
			ORDER BY a.creado_en DESC, a.id DESC LIMIT 300`, miembro, accion, tiendaID)
		if err != nil {
			return err
		}
		a.Registros, err = pgx.CollectRows(filas, func(f pgx.CollectableRow) (Registro, error) {
			var r Registro
			err := f.Scan(&r.ID, &r.Cuando, &r.Miembro.ID, &r.Miembro.Nombre, &r.Miembro.Rol, &r.Accion, &r.Texto, &r.TiendaID, &r.Tienda)
			return r, err
		})
		if err != nil {
			return err
		}
		a.Miembros, err = miembros(ctx, tx, true)
		return err
	})
	return a, err
}
