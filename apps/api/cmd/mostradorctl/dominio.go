package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cloudflare"
)

// dominio: el dominio propio de una tienda (el .com.ar que regala Grupo Vimo, registrado a su
// nombre en NIC.ar). Con el token de Cloudflare de MOSTRADOR_CLOUDFLARE_TOKEN y la cuenta de
// MOSTRADOR_CLOUDFLARE_CUENTA; el Worker es MOSTRADOR_WORKER (por defecto "mostrador").
//
//	regalar <slug> <dominio>  agrega la zona y el dominio a la tienda; dice qué nameservers cargar
//	activar <dominio>         con la zona activa, apunta el dominio (y su www) al Worker y lo deja
//	                          verificado y principal
//	quitar <dominio>          lo desapunta y lo saca de la tienda (la zona queda)
func dominio(ctx context.Context, base string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("uso: mostradorctl dominio regalar <slug> <dominio> | activar <dominio> | quitar <dominio>")
	}
	token, cuenta := os.Getenv("MOSTRADOR_CLOUDFLARE_TOKEN"), os.Getenv("MOSTRADOR_CLOUDFLARE_CUENTA")
	if token == "" || cuenta == "" {
		return fmt.Errorf("faltan MOSTRADOR_CLOUDFLARE_TOKEN y MOSTRADOR_CLOUDFLARE_CUENTA")
	}
	worker := os.Getenv("MOSTRADOR_WORKER")
	if worker == "" {
		worker = "mostrador"
	}
	cf := cloudflare.Nuevo(token, cuenta)
	conn, err := pgx.Connect(ctx, base)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	nombre := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(args[len(args)-1])), "www.")

	switch args[0] {
	case "regalar":
		if len(args) != 3 {
			return fmt.Errorf("uso: mostradorctl dominio regalar <slug> <dominio>")
		}
		zona, err := cf.ZonaDe(ctx, nombre)
		if err != nil {
			return err
		}
		if zona == nil {
			z, err := cf.CrearZona(ctx, nombre)
			if err != nil {
				return err
			}
			zona = &z
		}
		if err := enTienda(ctx, conn, args[1], func(tx pgx.Tx, tienda string) error {
			tag, err := tx.Exec(ctx, `INSERT INTO dominios (nombre, tienda_id) VALUES ($1, $2) ON CONFLICT (nombre) DO NOTHING`, nombre, tienda)
			if err == nil && tag.RowsAffected() == 0 {
				var ya int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM dominios WHERE nombre = $1`, nombre).Scan(&ya); err != nil || ya == 0 {
					return fmt.Errorf("%s ya es de otra tienda", nombre)
				}
			}
			return err
		}); err != nil {
			return err
		}
		fmt.Printf("%s agregado a %s. Zona de Cloudflare: %s\n", nombre, args[1], zona.Estado)
		if zona.Estado != "active" {
			fmt.Printf("Cargá estos nameservers en NIC.ar:\n  %s\nY cuando Cloudflare la dé por activa: mostradorctl dominio activar %s\n",
				strings.Join(zona.Nameservers, "\n  "), nombre)
		} else {
			fmt.Printf("La zona ya está activa: mostradorctl dominio activar %s\n", nombre)
		}
	case "activar":
		zona, err := cf.ZonaDe(ctx, nombre)
		if err != nil {
			return err
		}
		if zona == nil {
			return fmt.Errorf("%s no tiene zona en Cloudflare: primero mostradorctl dominio regalar", nombre)
		}
		if zona.Estado != "active" {
			return fmt.Errorf("la zona %s todavía está %q: ¿están cargados en NIC.ar los nameservers %s?",
				zona.Nombre, zona.Estado, strings.Join(zona.Nameservers, ", "))
		}
		hosts := []string{nombre}
		if nombre == zona.Nombre {
			hosts = append(hosts, "www."+nombre)
		}
		for _, h := range hosts {
			if err := cf.Apuntar(ctx, zona.ID, h, worker); err != nil {
				return fmt.Errorf("%s: %w", h, err)
			}
		}
		if err := enTiendaDeDominio(ctx, conn, nombre, func(tx pgx.Tx, tienda string) error {
			if _, err := tx.Exec(ctx, `UPDATE dominios SET verificado_en = coalesce(verificado_en, now()), principal = (nombre = $2)
				WHERE tienda_id = $1`, tienda, nombre); err != nil {
				return err
			}
			// El principal sólo entre los verificados: los demás quedan como estaban (sin principal).
			_, err := tx.Exec(ctx, `UPDATE dominios SET principal = false WHERE tienda_id = $1 AND verificado_en IS NULL`, tienda)
			return err
		}); err != nil {
			return err
		}
		fmt.Printf("Listo: %s muestra la tienda (en unos minutos, cuando Cloudflare emita el certificado).\n", strings.Join(hosts, " y "))
	case "quitar":
		zona, err := cf.ZonaDe(ctx, nombre)
		if err != nil {
			return err
		}
		if zona != nil {
			for _, h := range []string{nombre, "www." + nombre} {
				if err := cf.Desapuntar(ctx, zona.ID, h); err != nil {
					return err
				}
			}
		}
		if err := enTiendaDeDominio(ctx, conn, nombre, func(tx pgx.Tx, tienda string) error {
			_, err := tx.Exec(ctx, `DELETE FROM dominios WHERE tienda_id = $1 AND nombre = $2`, tienda, nombre)
			return err
		}); err != nil {
			return err
		}
		fmt.Println("Quitado:", nombre)
	default:
		return fmt.Errorf("dominio: subcomando %q desconocido", args[0])
	}
	return nil
}

// enTienda corre fn en una transacción con la tienda de ese slug (el migrador ve sólo esa, por RLS).
func enTienda(ctx context.Context, conn *pgx.Conn, slug string, fn func(pgx.Tx, string) error) error {
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		var id *string
		if err := tx.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id); err != nil {
			return err
		}
		if id == nil {
			return fmt.Errorf("no hay una tienda abierta con el slug %q", slug)
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.tienda_id', $1, true)`, *id); err != nil {
			return err
		}
		return fn(tx, *id)
	})
}

// enTiendaDeDominio: como enTienda, con la tienda dueña del dominio.
func enTiendaDeDominio(ctx context.Context, conn *pgx.Conn, nombre string, fn func(pgx.Tx, string) error) error {
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		var id string
		if _, err := tx.Exec(ctx, `SELECT set_config('app.dominio', $1, true)`, nombre); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `SELECT tienda_id::text FROM dominios WHERE nombre = $1`, nombre).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s no está agregado a ninguna tienda", nombre)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.dominio', '', true), set_config('app.tienda_id', $1, true)`, id); err != nil {
			return err
		}
		return fn(tx, id)
	})
}
