import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const browser = await chromium.launch();
const log = [];
const c = await ctx(browser, true); const page = await c.newPage(); watch(page, log);
await page.goto(BASE + '/ayuda', { waitUntil: 'networkidle' }); await cookies(page);
const cats = await page.evaluate(() => [...new Set([...document.querySelectorAll('a[href^="/ayuda/"]')].map(a => a.getAttribute('href')))]);
console.log('ayuda links', cats.length, cats.join(' '));
// search
const s = page.locator('input[type=search], input[placeholder*="usc" i]').first();
if (await s.count()) { await s.fill('envío'); await page.waitForTimeout(1500); await page.screenshot({ path: SS + 'm_ayuda_search.png' });
  console.log('search results', (await page.evaluate(() => document.querySelector('main')?.innerText)).replace(/\n+/g,' | ').slice(0, 700));
  await s.press('Enter'); await page.waitForTimeout(1500); console.log('after enter url', page.url());
  await s.fill('xyzqwe'); await page.waitForTimeout(1200); console.log('nores:', (await page.evaluate(() => document.querySelector('main')?.innerText)).replace(/\n+/g,' | ').slice(0, 300));
} else console.log('no search input');
const arts = cats.filter(h => h.split('/').length >= 4);
const all = new Set(arts);
for (const cat of cats.filter(h => h.split('/').length === 3).slice(0, 20)) {
  const r = await page.goto(BASE + cat, { waitUntil: 'networkidle' }); if (r.status() >= 400) console.log('CAT', r.status(), cat);
  (await page.evaluate(() => [...document.querySelectorAll('a[href^="/ayuda/"]')].map(a => a.getAttribute('href')))).filter(h => h.split('/').length >= 4).forEach(h => all.add(h));
}
console.log('articles', all.size);
let i = 0; const bad = [];
for (const a of all) {
  const r = await page.goto(BASE + a, { waitUntil: 'domcontentloaded' });
  const of = await overflow(page);
  const t = await page.evaluate(() => document.querySelector('main')?.innerText || '');
  if (r.status() >= 400 || of.sw > of.vw || /mostrador/i.test(t) || t.length < 200) bad.push(`${a} st=${r.status()} sw=${of.sw} len=${t.length} most=${/mostrador/i.test(t)}`);
  if (i < 3) await page.screenshot({ path: SS + `m_ayuda_art_${i}.png`, fullPage: true });
  i++;
}
console.log('bad articles', bad);
// Panel/no-JS mention of features that don't exist? sample first article text
await page.goto(BASE + [...all][0], { waitUntil: 'networkidle' });
console.log('sample', (await page.evaluate(() => document.querySelector('main')?.innerText)).slice(0, 1200));
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
