import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const browser = await chromium.launch();
const log = [];
const c = await ctx(browser, true); const page = await c.newPage(); watch(page, log);
for (const s of ['terminos','privacidad','cookies','arrepentimiento','reembolsos','defensa-del-consumidor','tratamiento-de-datos']) {
  const r = await page.goto(BASE + '/legales/' + s, { waitUntil: 'networkidle' }); const of = await overflow(page);
  const t = await page.evaluate(() => document.querySelector('main')?.innerText || '');
  console.log(s, r.status(), of.sw > of.vw ? 'OVERFLOW' : '', t.length, /mostrador/i.test(t) ? 'MOSTRADOR' : '', (t.match(/\[.*?\]|completar|TODO|lorem/gi)||[]).slice(0,3));
}
await page.goto(BASE + '/contacto', { waitUntil: 'networkidle' }); await cookies(page);
const fields = await page.evaluate(() => [...document.querySelectorAll('form input, form textarea, form select')].map(e => `${e.tagName}:${e.name}:${e.type}:${e.required}:${e.placeholder}`));
console.log('fields', fields);
await page.screenshot({ path: SS + 'm_contacto_form.png', fullPage: true });
// submit empty first
const submit = page.locator('form button[type=submit], form button:not([type])').last();
console.log('submit text', await submit.innerText());
await submit.click(); await page.waitForTimeout(1000);
console.log('empty submit url', page.url(), (await page.evaluate(() => [...document.querySelectorAll('[role=alert], .error, [aria-invalid=true]')].map(e => e.innerText || e.name).join('; '))));
for (const f of await page.locator('form input, form textarea').all()) {
  const n = (await f.getAttribute('name')) || ''; const t = (await f.getAttribute('type')) || '';
  if (t === 'hidden' || t === 'checkbox' || t === 'radio') continue;
  if (/mail/i.test(n + t)) await f.fill('nahuel+qa1@grupovimo.com.ar');
  else if (/tel|whats|phone/i.test(n + t)) await f.fill('3415550000');
  else if (/nombre|name/i.test(n)) await f.fill('QA Pruebas');
  else if (/url|web|tienda|sitio/i.test(n + t)) await f.fill('https://ejemplo.com.ar');
  else if (/mensaje|consulta|comentario/i.test(n) || (await f.evaluate(e => e.tagName)) === 'TEXTAREA') await f.fill('MENSAJE DE PRUEBA QA (automatizado). Ignorar, por favor. No requiere respuesta.');
  else await f.fill('QA prueba');
}
for (const sel of await page.locator('form select').all()) { const opts = await sel.locator('option').allInnerTexts(); await sel.selectOption({ index: Math.min(1, opts.length - 1) }); }
const radios = page.locator('form input[type=radio]'); if (await radios.count()) await radios.first().check({ force: true });
await page.screenshot({ path: SS + 'm_contacto_filled.png', fullPage: true });
await submit.click(); await page.waitForTimeout(4000);
await page.screenshot({ path: SS + 'm_contacto_sent.png', fullPage: true });
console.log('after submit', page.url(), (await page.evaluate(() => document.querySelector('main')?.innerText)).replace(/\n+/g,' | ').slice(0, 600));
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
