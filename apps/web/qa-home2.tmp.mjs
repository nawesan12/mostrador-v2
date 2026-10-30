import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const browser = await chromium.launch();
const log = [];
for (const mobile of [true, false]) {
  const c = await ctx(browser, mobile); const page = await c.newPage(); watch(page, log);
  const tag = mobile ? 'm' : 'd';
  await page.goto(BASE + '/', { waitUntil: 'networkidle' }); await cookies(page);
  // Ver plantillas
  await page.getByRole('link', { name: /Ver plantillas/ }).or(page.getByRole('button', { name: /Ver plantillas/ })).first().click();
  await page.waitForTimeout(1500);
  console.log(tag, 'after Ver plantillas url', page.url());
  await page.screenshot({ path: `${SS}${tag}_verplantillas.png` });
  const txt = await page.evaluate(() => document.body.innerText.slice(0, 1500));
  console.log(txt.replace(/\n+/g, ' | ').slice(0, 900));
  await page.goto(BASE + '/', { waitUntil: 'networkidle' });
  // pricing toggle
  await page.getByText('Anual', { exact: false }).first().click().catch(e => console.log('anual click fail', e.message.slice(0,80)));
  await page.waitForTimeout(600);
  const prices = await page.evaluate(() => [...document.body.innerText.matchAll(/\$ ?[\d.]+ ?\/ ?(mes|año)[^\n]*/g)].map(m=>m[0]));
  console.log(tag, 'annual prices', prices, (document => 0)());
  const ptxt = await page.evaluate(() => { const el=[...document.querySelectorAll('section')].find(s=>/Más barato en cada/.test(s.innerText)); return el?.innerText.replace(/\n+/g,' | ').slice(0,1200); });
  console.log(ptxt);
  await page.locator('#precios').screenshot({ path: `${SS}${tag}_precios_anual.png` }).catch(()=>{});
  if (mobile) {
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.getByRole('button', { name: /menú|menu|abrir/i }).first().click().catch(e => console.log('menu fail', e.message.slice(0,80)));
    await page.waitForTimeout(700); await page.screenshot({ path: `${SS}m_menu.png` });
    console.log('menu text', (await page.evaluate(() => document.body.innerText.slice(0,600))).replace(/\n+/g,' | '));
    await page.keyboard.press('Escape');
  }
  // tincho fab
  await page.goto(BASE + '/', { waitUntil: 'networkidle' });
  const fab = page.locator('button[aria-label*="Tincho" i], a[aria-label*="Tincho" i]').last();
  console.log('fab count', await page.locator('button[aria-label*="Tincho" i], a[aria-label*="Tincho" i]').count());
  await fab.click().catch(e => console.log('fab fail', e.message.slice(0, 80)));
  await page.waitForTimeout(1200); await page.screenshot({ path: `${SS}${tag}_tincho_fab.png` });
  const inp = page.locator('textarea, input[type=text]').last();
  if (await inp.count()) { await inp.fill('¿Cuánto cuesta el plan Emprende y tiene comisión por venta?'); await inp.press('Enter'); await page.waitForTimeout(12000);
    await page.screenshot({ path: `${SS}${tag}_tincho_fab_resp.png` });
    console.log('tincho resp', (await page.evaluate(() => document.body.innerText.slice(-1500))).replace(/\n+/g,' | ')); }
  await c.close();
}
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
