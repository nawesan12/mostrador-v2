// Package db es el único lugar que habla con el pool de Postgres. Todo acceso a datos de una tienda
// pasa por EnTienda, que abre una transacción y fija la tienda con set_config(..., true): el valor
// vive sólo en esa transacción, así que un pool compartido nunca filtra una tienda a otra.
// Un olvido devuelve vacío (RLS), nunca datos ajenos.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool es la conexión de un rol (app, worker o staff).
type Pool struct{ *pgxpool.Pool }

// Abrir conecta con la URL de un rol y verifica que responda.
func Abrir(ctx context.Context, url string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("url de la base: %w", err)
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, fmt.Errorf("la base no responde: %w", err)
	}
	return &Pool{p}, nil
}

// ErrSinTienda: se pidió un acceso por tienda sin tienda.
var ErrSinTienda = errors.New("acceso a datos sin tienda")

// EnTienda corre fn en una transacción con la tienda fijada. Si fn devuelve error, se revierte.
func (p *Pool) EnTienda(ctx context.Context, tiendaID string, fn func(pgx.Tx) error) error {
	if tiendaID == "" {
		return ErrSinTienda
	}
	return p.enTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tienda_id', $1, true)", tiendaID); err != nil {
			return err
		}
		return fn(tx)
	})
}

// EnTx corre fn en una transacción sin tienda: sólo sirve para lo que no depende de una
// (alta de tienda con app.crear_tienda, el worker sobre el outbox, lecturas de staff).
func (p *Pool) EnTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return p.enTx(ctx, fn)
}

func (p *Pool) enTx(ctx context.Context, fn func(pgx.Tx) error) (err error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
