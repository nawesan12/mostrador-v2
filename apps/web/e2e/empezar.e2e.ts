import { expect, test } from '@playwright/test';

test('/empezar anota el origen (una sola vez) y lleva a crear la cuenta', async ({
	page,
	context
}) => {
	await page.goto('/empezar?utm_source=flyer&ref=BRUMA-REF&cupon=arranco50');
	await expect(page).toHaveURL(/\/ingresar\?m=registro$/);
	const leer = async () =>
		JSON.parse(
			decodeURIComponent(
				(await context.cookies()).find((c) => c.name === 'mostrador-origen')!.value
			)
		);
	expect(await leer()).toMatchObject({ fuente: 'flyer', ref: 'BRUMA-REF', cupon: 'ARRANCO50' });
	// Una visita posterior con otro ref no pisa el primero.
	await page.goto('/empezar?ref=OTRO');
	expect((await leer()).ref).toBe('BRUMA-REF');
	// Basura no se anota.
	await context.clearCookies();
	await page.goto('/empezar?ref=%3Cscript%3E');
	expect((await context.cookies()).some((c) => c.name === 'mostrador-origen')).toBe(false);
});
