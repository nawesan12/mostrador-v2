package ia

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Base es lo que se usa del pool (db.Pool).
type Base interface {
	EnTienda(ctx context.Context, tiendaID string, fn func(pgx.Tx) error) error
	EnTx(ctx context.Context, fn func(pgx.Tx) error) error
}

// Rutas de ia_uso. «motor» son las respuestas sin IA (costo cero): se cuentan para el plan.
const (
	RutaChat     = "chat"
	RutaMejorar  = "mejorar"
	RutaTelegram = "telegram"
	RutaMotor    = "motor"
)

var argentina = time.FixedZone("ART", -3*3600)

// Dia y Mes en hora de Argentina.
func Dia(t time.Time) string { return t.In(argentina).Format(time.DateOnly) }
func Mes(t time.Time) string { return t.In(argentina).Format("2006-01") }
func inicioDeMes(t time.Time) string {
	a := t.In(argentina)
	return time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, argentina).Format(time.DateOnly)
}

// Registrar suma una respuesta (con sus tokens y su costo) al día de la tienda y al gasto del mes
// de la plataforma. Se llama después del modelo, en una transacción corta.
func Registrar(ctx context.Context, b Base, tiendaID, ruta string, u Uso, ahora time.Time) error {
	costo := u.Costo()
	return b.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `INSERT INTO ia_uso (tienda_id, dia, ruta, pedidos, entrada, cache_leido, cache_escrito, salida, costo)
			VALUES ($1, $2::date, $3, 1, $4, $5, $6, $7, $8)
			ON CONFLICT (tienda_id, dia, ruta) DO UPDATE SET pedidos = ia_uso.pedidos + 1,
				entrada = ia_uso.entrada + EXCLUDED.entrada, cache_leido = ia_uso.cache_leido + EXCLUDED.cache_leido,
				cache_escrito = ia_uso.cache_escrito + EXCLUDED.cache_escrito, salida = ia_uso.salida + EXCLUDED.salida,
				costo = ia_uso.costo + EXCLUDED.costo`,
			tiendaID, Dia(ahora), ruta, u.Entrada, u.CacheLeido, u.CacheEscrito, u.Salida, costo); err != nil {
			return err
		}
		if ruta == RutaMotor {
			return nil
		}
		_, err := t.Exec(ctx, `INSERT INTO app.ia_gasto (mes, costo, pedidos) VALUES ($1, $2, 1)
			ON CONFLICT (mes) DO UPDATE SET costo = app.ia_gasto.costo + EXCLUDED.costo, pedidos = app.ia_gasto.pedidos + 1`,
			Mes(ahora), costo)
		return err
	})
}

// Consumo de una tienda: respuestas con IA del mes y de hoy, y las del motor (sin IA) del mes.
type Consumo struct {
	Mes   int `json:"mes"`
	Hoy   int `json:"hoy"`
	SinIA int `json:"sinIa"`
}

// ConsumoEnTx lee el consumo dentro de una transacción ya abierta en la tienda.
func ConsumoEnTx(ctx context.Context, t pgx.Tx, tiendaID string, ahora time.Time) (Consumo, error) {
	var c Consumo
	err := t.QueryRow(ctx, `SELECT
			coalesce(sum(pedidos) FILTER (WHERE ruta <> 'motor'), 0),
			coalesce(sum(pedidos) FILTER (WHERE ruta <> 'motor' AND dia = $3::date), 0),
			coalesce(sum(pedidos) FILTER (WHERE ruta = 'motor'), 0)
		FROM ia_uso WHERE tienda_id = $1 AND dia >= $2::date`, tiendaID, inicioDeMes(ahora), Dia(ahora)).
		Scan(&c.Mes, &c.Hoy, &c.SinIA)
	return c, err
}

// Gastado es el gasto de la plataforma en el mes, en micro-dólares.
func Gastado(ctx context.Context, b Base, ahora time.Time) (int64, error) {
	var n int64
	err := b.EnTx(ctx, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `SELECT costo FROM app.ia_gasto WHERE mes = $1`, Mes(ahora)).Scan(&n)
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	})
	return n, err
}

// Topes de una tienda: respuestas con IA por mes (las del plan) y por día; Presupuesto es la parte
// de la plataforma en micro-dólares por mes (0: sin tope).
type Topes struct {
	Mes, Dia    int
	Presupuesto int64
}

// Corte dice por qué no se llama al modelo ("" si se puede).
type Corte string

const (
	SinCorte        Corte = ""
	CortePlan       Corte = "plan"
	CorteDia        Corte = "dia"
	CortePlataforma Corte = "plataforma"
)

// Decidir compara el consumo con los topes.
func Decidir(c Consumo, gastado int64, t Topes) Corte {
	switch {
	case t.Presupuesto > 0 && gastado >= t.Presupuesto:
		return CortePlataforma
	case c.Mes >= t.Mes:
		return CortePlan
	case c.Hoy >= t.Dia:
		return CorteDia
	}
	return SinCorte
}

// Revisar lee el consumo de la tienda y el gasto del mes y dice si se puede llamar al modelo.
// Dos preguntas a la vez pueden pasarse por una: no vale una transacción abierta durante la llamada.
func Revisar(ctx context.Context, b Base, tiendaID string, t Topes, ahora time.Time) (Corte, Consumo, error) {
	var c Consumo
	if err := b.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		var err error
		c, err = ConsumoEnTx(ctx, tx, tiendaID, ahora)
		return err
	}); err != nil {
		return SinCorte, c, err
	}
	gastado, err := Gastado(ctx, b, ahora)
	if err != nil {
		return SinCorte, c, err
	}
	return Decidir(c, gastado, t), c, nil
}
