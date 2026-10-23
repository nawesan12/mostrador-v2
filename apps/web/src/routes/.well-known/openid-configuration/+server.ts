import { json, preflight } from '#lib/server/mcp/http.ts';
import { metadatosServidor } from '#lib/server/mcp/oauth.ts';
import type { RequestHandler } from './$types';

/** Algunos clientes MCP buscan acá antes que en oauth-authorization-server: mismos metadatos. */
export const GET: RequestHandler = ({ url }) => json(metadatosServidor(url.origin));
export const OPTIONS = preflight;
