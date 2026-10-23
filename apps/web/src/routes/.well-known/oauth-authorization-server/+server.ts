import { json, preflight } from '#lib/server/mcp/http.ts';
import { metadatosServidor } from '#lib/server/mcp/oauth.ts';
import type { RequestHandler } from './$types';

/** Metadatos del servidor de autorización (RFC 8414). */
export const GET: RequestHandler = ({ url }) => json(metadatosServidor(url.origin));
export const OPTIONS = preflight;
