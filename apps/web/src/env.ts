import { defineEnvVars } from '@sveltejs/kit/env';

// Los secretos y bindings se leen de `cloudflare:workers`, como en Caja.
export const variables = defineEnvVars({});
