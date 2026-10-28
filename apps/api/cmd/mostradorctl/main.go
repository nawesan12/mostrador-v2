// Comando mostradorctl: tareas de operación de la base.
//
//	mostradorctl bootstrap   crea roles y base (superusuario: MOSTRADOR_BASE_DE_DATOS_ADMIN)
//	mostradorctl migrar      aplica las migraciones (MOSTRADOR_BASE_DE_DATOS_MIGRADOR)
//	mostradorctl partner-demo <email>
//	                         demo del panel de partners para una cuenta que ya existe (sólo local:
//	                         superusuario de MOSTRADOR_BASE_DE_DATOS_ADMIN sobre la base del migrador)
//
// Las claves de los roles salen de MOSTRADOR_CLAVE_<ROL> (vacías en desarrollo).
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/partners"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: mostradorctl bootstrap|migrar|partner-demo <email>")
		os.Exit(2)
	}
	if err := correr(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func correr(cmd string) error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch cmd {
	case "bootstrap":
		if err := cfg.Exigir("BaseAdmin", "BaseMigrador"); err != nil {
			return err
		}
		u, err := url.Parse(cfg.BaseMigrador)
		if err != nil {
			return err
		}
		claves := db.Claves{}
		for _, r := range db.Roles {
			claves[r] = os.Getenv("MOSTRADOR_CLAVE_" + strings.ToUpper(strings.TrimPrefix(r, "mostrador_")))
		}
		if err := db.Bootstrap(ctx, cfg.BaseAdmin, strings.TrimPrefix(u.Path, "/"), claves); err != nil {
			return err
		}
		fmt.Println("roles y base listos")
	case "migrar":
		if err := cfg.Exigir("BaseMigrador"); err != nil {
			return err
		}
		if err := db.Migrar(ctx, cfg.BaseMigrador); err != nil {
			return err
		}
		fmt.Println("migraciones aplicadas")
	case "partner-demo":
		if len(os.Args) < 3 {
			return fmt.Errorf("falta el email: mostradorctl partner-demo nombre@mail.com")
		}
		if !cfg.Desarrollo() {
			return fmt.Errorf("la demo de partners es sólo para desarrollo")
		}
		if err := cfg.Exigir("BaseAdmin", "BaseMigrador"); err != nil {
			return err
		}
		admin, err := url.Parse(cfg.BaseAdmin)
		if err != nil {
			return err
		}
		mig, err := url.Parse(cfg.BaseMigrador)
		if err != nil {
			return err
		}
		admin.Path = mig.Path
		conn, err := pgx.Connect(ctx, admin.String())
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		if _, err := partners.Sembrar(ctx, conn, os.Args[2], time.Now()); err != nil {
			return err
		}
		fmt.Println("demo de partners lista: /partners/panel con", os.Args[2])
	default:
		return fmt.Errorf("comando desconocido %q", cmd)
	}
	return nil
}
