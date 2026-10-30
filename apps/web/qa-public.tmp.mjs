import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const paths = process.argv.slice(2);
const browser = await chromium.launch();
const log = [];
for (const mobile of [false, true]) {
  const c = await ctx(browser, mobile); const page = await c.newPage(); watch(page, log);
  let first = true;
  for (const p of paths) {
    const r = await page.goto(BASE + p, { waitUntil: 'networkidle', timeout: 45000 }).catch(e => ({ status: () => 'ERR ' + e.message.slice(0,80) }));
    if (first) { await page.waitForTimeout(800); const ck = await cookies(page); log.push(`cookie banner: ${ck}`); first = false; }
    const of = await overflow(page);
    const info = await page.evaluate(() => ({
      title: document.title, desc: document.querySelector('meta[name=description]')?.content, og: document.querySelector('meta[property="og:image"]')?.content, ogt: document.querySelector('meta[property="og:title"]')?.content,
      mostrador: (document.body.innerText.match(/.{0,40}mostrador.{0,40}/gi) || []).slice(0,3),
      h1: [...document.querySelectorAll('h1')].map(h=>h.innerText).join(' | ').slice(0,150),
    }));
    const name = (mobile ? 'm' : 'd') + '_' + (p.replace(/[^a-z0-9]+/gi, '_') || 'home') + '.png';
    await page.screenshot({ path: SS + name, fullPage: true });
    console.log(JSON.stringify({ p, mobile, status: r?.status?.(), ...info, of: of.sw > of.vw ? of : 'ok', shot: name }));
  }
  await c.close();
}
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
