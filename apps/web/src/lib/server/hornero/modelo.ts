/**
 * Hornero con Workers AI: manda la conversación con el contexto de la tienda y devuelve el texto a
 * medida que llega (streaming). Acepta los dos formatos de stream de Workers AI.
 */
export const MODELO = '@cf/google/gemma-4-26b-a4b-it';

export type Mensaje = { role: 'system' | 'user' | 'assistant'; content: string };

/** Pasa el stream SSE de Workers AI a texto plano. */
export function soloTexto(sse: ReadableStream<Uint8Array>) {
	const dec = new TextDecoder();
	const enc = new TextEncoder();
	let resto = '';
	return sse.pipeThrough(
		new TransformStream<Uint8Array, Uint8Array>({
			transform(trozo, ctl) {
				resto += dec.decode(trozo, { stream: true });
				const lineas = resto.split('\n');
				resto = lineas.pop() ?? '';
				for (const l of lineas) {
					const dato = l.startsWith('data:') ? l.slice(5).trim() : '';
					if (!dato || dato === '[DONE]') continue;
					try {
						const j = JSON.parse(dato) as {
							response?: string;
							choices?: { delta?: { content?: string | null } }[];
						};
						const t = j.response ?? j.choices?.[0]?.delta?.content ?? '';
						if (t) ctl.enqueue(enc.encode(t));
					} catch {
						// Línea partida o metadatos: se ignora.
					}
				}
			}
		})
	);
}

export async function responderConIa(ai: Ai, mensajes: Mensaje[]) {
	const sse = (await ai.run(
		MODELO as Parameters<Ai['run']>[0],
		{
			messages: mensajes,
			stream: true,
			max_tokens: 500,
			temperature: 0.4
		} as never
	)) as unknown as ReadableStream<Uint8Array>;
	return soloTexto(sse);
}
