#!/bin/sh
# Despliega Tu Local Web en el VPS (todo corre ahí; Cloudflare es sólo DNS).
#   scripts/desplegar-vps.sh          la web y la API
#   scripts/desplegar-vps.sh web      sólo la web
#   scripts/desplegar-vps.sh api      sólo la API (con sus migraciones)
set -eu
VPS=root@187.41.102.190
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
QUE=${1:-todo}

if [ "$QUE" = todo ] || [ "$QUE" = api ]; then
	SAL=$(mktemp -d)
	for c in api worker mostradorctl; do
		(cd "$RAIZ/apps/api" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$SAL/$c" "./cmd/$c")
		scp -q "$SAL/$c" "$VPS:/opt/tulocalweb/bin/$c.nuevo"
	done
	rm -rf "$SAL"
	ssh "$VPS" 'cd /opt/tulocalweb/bin && for c in api worker mostradorctl; do mv $c.nuevo $c; done &&
		cd .. && docker compose run --rm --no-deps -T api /app/mostradorctl migrar && docker compose up -d --force-recreate --no-deps api worker'
fi

if [ "$QUE" = todo ] || [ "$QUE" = web ]; then
	(cd "$RAIZ/apps/web" && rm -rf build && DESTINO=vps bun run build >/dev/null)
	rsync -a --delete "$RAIZ/apps/web/build/" "$VPS:/opt/tulocalweb/web/build/"
	rsync -a --delete "$RAIZ/apps/web/migraciones-sandbox/" "$VPS:/opt/tulocalweb/web/migraciones/sandbox/"
	rsync -a --delete "$RAIZ/apps/web/migraciones-estado/" "$VPS:/opt/tulocalweb/web/migraciones/estado/"
	ssh "$VPS" 'cd /opt/tulocalweb && docker compose up -d --force-recreate --no-deps web'
fi

sleep 4
curl -s -o /dev/null -w "tulocalweb.com.ar %{http_code}\n" https://tulocalweb.com.ar/
