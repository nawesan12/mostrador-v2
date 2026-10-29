package importacion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
)

// empretienda lee la tienda pública: el sitemap da las fichas y cada paso lee 20. El cursor es la
// posición en el sitemap (la lista entera no entra en los 2 KB del cursor), así que cada paso lo
// vuelve a pedir: es un solo pedido y lo que cambie en el medio se toma igual.
type empretienda struct {
	http   *http.Client
	origen string
}

const porPasoET = 20

type cursorET struct {
	Desde int `json:"desde"`
}

func (e empretienda) Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error) {
	if etapa != "productos" {
		return Pagina{}, nil
	}
	var c cursorET
	_ = json.Unmarshal(cursor, &c)
	xml, _, err := importar.Pedir(ctx, e.http, e.origen+"/sitemap.xml")
	if err != nil {
		return Pagina{}, definitivo(err)
	}
	fichas := importar.FichasDelSitemap(string(xml), e.origen)
	if c.Desde >= len(fichas) {
		return Pagina{}, nil
	}
	tanda := fichas[c.Desde:min(c.Desde+porPasoET, len(fichas))]
	// Las fichas se leen de a 4 a la vez: rápido sin cargar la tienda.
	leidas := make([]*importar.Importado, len(tanda))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	var mu sync.Mutex
	var primerError error
	for i, u := range tanda {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			html, _, err := importar.Pedir(ctx, e.http, u)
			if err != nil {
				// Una ficha que no responde se saltea; si fallan todas, se reintenta el paso.
				mu.Lock()
				primerError = err
				mu.Unlock()
				return
			}
			if f, ok := importar.FichaDeHTML(string(html), u); ok {
				if p, ok := importar.DeEmpretienda(f); ok {
					leidas[i] = &p
				}
			}
		}()
	}
	wg.Wait()
	var items []importar.Importado
	for _, p := range leidas {
		if p != nil {
			items = append(items, *p)
		}
	}
	if len(items) == 0 && primerError != nil {
		return Pagina{}, fmt.Errorf("importacion: no respondió ninguna ficha de Empretienda: %w", primerError)
	}
	p := Pagina{Productos: productosDe(items)}
	if sig := c.Desde + len(tanda); sig < len(fichas) {
		p.Siguiente, _ = json.Marshal(cursorET{Desde: sig})
	}
	return p, nil
}
