import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const browser = await chromium.launch();
const log = [];
const slugs = process.argv.slice(2);
for (const mobile of [true, false]) {
  const c = await ctx(browser, mobile); const page = await c.newPage(); watch(page, log);
  const tag = mobile ? 'm' : 'd';
  for (const s of slugs) {
    const r = await page.goto(`${BASE}/t/${s}`, { waitUntil: 'networkidle' }).catch(e => null);
    console.log(tag, s, 'status', r?.status(), 'title', await page.title());
    if (!r || r.status() >= 400) continue;
    await cookies(page);
    await page.screenshot({ path: `${SS}${tag}_t_${s}.png` });
    const of = await overflow(page); if (of.sw > of.vw) console.log('  OVERFLOW', JSON.stringify(of));
    const prodLinks = await page.evaluate(() => [...new Set([...document.querySelectorAll('a[href*="/producto/"]')].map(a => a.getAttribute('href')))]);
    console.log('  products linked', prodLinks.length, prodLinks.slice(0,3));
    const navLinks = await page.evaluate(() => [...new Set([...document.querySelectorAll('a[href]')].map(a => a.getAttribute('href')))].filter(h => !h.includes('/producto/')));
    console.log('  other links', navLinks.join(' '));
    if (prodLinks[0]) {
      await page.goto(new URL(prodLinks[0], BASE).href, { waitUntil: 'networkidle' });
      await page.screenshot({ path: `${SS}${tag}_t_${s}_prod.png`, fullPage: true });
      const of2 = await overflow(page); if (of2.sw > of2.vw) console.log('  PROD OVERFLOW', JSON.stringify(of2));
      const btn = page.getByRole('button', { name: /agregar|sumar al carrito|comprar/i }).first();
      if (await btn.count()) { await btn.click().catch(e => console.log('  add fail', e.message.slice(0,100))); await page.waitForTimeout(1500);
        await page.screenshot({ path: `${SS}${tag}_t_${s}_added.png` }); }
      else console.log('  no add button; buttons:', await page.evaluate(() => [...document.querySelectorAll('button')].map(b=>b.innerText.trim()).filter(Boolean).slice(0,15).join(' / ')));
      await page.goto(`${BASE}/t/${s}/carrito`, { waitUntil: 'networkidle' });
      await page.screenshot({ path: `${SS}${tag}_t_${s}_cart.png`, fullPage: true });
      console.log('  cart:', (await page.evaluate(() => document.querySelector('main')?.innerText || document.body.innerText)).replace(/\n+/g, ' | ').slice(0, 400));
    }
    for (const sub of ['catalogo', 'nosotros', 'arrepentimiento']) {
      const r2 = await page.goto(`${BASE}/t/${s}/${sub}`, { waitUntil: 'networkidle' });
      const of3 = await overflow(page);
      console.log('  ', sub, r2.status(), of3.sw > of3.vw ? 'OVERFLOW ' + JSON.stringify(of3) : '');
      if (s === slugs[0]) await page.screenshot({ path: `${SS}${tag}_t_${s}_${sub}.png`, fullPage: true });
    }
  }
  await c.close();
}
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
