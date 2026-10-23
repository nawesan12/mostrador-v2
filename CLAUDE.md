# Mostrador v2

Leé `docs/CONVENCIONES.md` y `docs/decisiones.md` antes de tocar código.

- Sigue la impronta de Caja (`~/Projects/caja`): antes de escribir algo, buscalo ahí y reutilizalo.
- Gestor: bun. Front siempre en Workers; API Go para el VPS; demo con `DATA_SOURCE=sandbox`.
- Una rama por PR (`pr-NN-tema`), CI verde, merge a `main`. Sin previews ni servers de desarrollo
  en Cloudflare: se prueba en local (`wrangler dev --local`) y sólo se despliega producción.
- Cloudflare con cuidado: nada de crons, polling ni scripts en masa contra producción.
- `CREDENCIALES.md` nunca al repo.
