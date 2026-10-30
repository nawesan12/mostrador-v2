import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
import fs from 'fs';
const browser = await chromium.launch();
const log = [];
const allLinks = new Set();
const pages = process.argv.slice(2);
for (const mobile of [true, false]) {
  const c = await ctx(browser, mobile); const page = await c.newPage(); watch(page, log);
  for (const p of pages) {
  await page.goto(BASE + p, { waitUntil: 'networkidle' }); await cookies(page);
  const h = await page.evaluate(() => document.documentElement.scrollHeight);
  let i = 0;
  for (let y = 0; y < h; y += 812) { await page.evaluate(y => window.scrollTo(0, y), y); await page.waitForTimeout(450);
    if (mobile) { await page.screenshot({ path: `${SS}seg_${p.replace(/\W+/g,'_')}_m_${String(i).padStart(2,'0')}.png` }); } i++; }
  await page.evaluate(() => window.scrollTo(0, 0)); await page.waitForTimeout(300);
  if (!mobile) await page.screenshot({ path: `${SS}full_${p.replace(/\W+/g,'_')}_d.png`, fullPage: true });
  const links = await page.evaluate(() => [...document.querySelectorAll('a[href]')].map(a => a.href));
  links.forEach(l => allLinks.add(l));
  if (!mobile) fs.writeFileSync(`${SS}text_${p.replace(/\W+/g,'_')}.txt`, await page.evaluate(() => document.body.innerText));
  console.log(p, mobile, 'segments', i, JSON.stringify(await overflow(page)));
  }
  await c.close();
}
// check links
const req = await (await chromium.launch()).newContext();
for (const l of allLinks) {
  if (!/^https?:/.test(l)) continue;
  try { const r = await req.request.get(l, { maxRedirects: 5, timeout: 20000 }); if (r.status() >= 400) console.log('BROKEN', r.status(), l); } catch (e) { console.log('LINKERR', l, e.message.slice(0, 80)); }
}
console.log('links checked', allLinks.size);
console.log([...allLinks].filter(l=>l.startsWith(BASE)).map(l=>l.replace(BASE,'')).sort().join(' '));
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
process.exit(0);
