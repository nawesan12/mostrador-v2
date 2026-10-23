import { json, preflight } from '#lib/server/mcp/http.ts';
import { metadatosRecurso } from '#lib/server/mcp/oauth.ts';
import type { RequestHandler } from './$types';

/** Con o sin la ruta del recurso (`/.well-known/oauth-protected-resource/api/mcp`, RFC 9728). */
export const GET: RequestHandler = ({ url }) => json(metadatosRecurso(url.origin));
export const OPTIONS = preflight;
