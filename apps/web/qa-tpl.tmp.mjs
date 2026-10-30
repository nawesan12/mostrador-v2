import { chromium, ctx, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
const browser = await chromium.launch();
const log = [];
const c = await ctx(browser, false); const page = await c.newPage(); watch(page, log);
await page.goto(BASE + '/', { waitUntil: 'networkidle' }); await cookies(page);
await page.getByRole('button', { name: /Ver plantillas/ }).first().click(); await page.waitForTimeout(1500);
const dlg = page.locator('[role=dialog], dialog').first();
console.log('dialog html links', await page.evaluate(() => [...document.querySelectorAll('[role=dialog] a, dialog a')].map(a => a.innerText.trim() + ' -> ' + a.href)));
console.log('iframes', await page.evaluate(() => [...document.querySelectorAll('iframe')].map(f => f.src)));
console.log('buttons', await page.evaluate(() => [...document.querySelectorAll('[role=dialog] button, dialog button')].map(b => (b.getAttribute('aria-label') || b.innerText).trim())));
await page.screenshot({ path: SS + 'd_tpl_0.png' });
const names = [];
for (let i = 0; i < 25; i++) {
  const n = await page.evaluate(() => { const d = document.querySelector("[role=dialog], dialog"); return d?.innerText.split("\n").filter(Boolean).slice(0,12).join("/"); });
  names.push(n);
  const nx = page.locator('[role=dialog] button[aria-label*="iguiente" i], dialog button[aria-label*="iguiente" i]');
  if (!(await nx.count())) { console.log('no next btn'); break; }
  await nx.first().click(); await page.waitForTimeout(700);
  if (i === 2) await page.screenshot({ path: SS + 'd_tpl_3.png' });
}
console.log('names', names);
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await browser.close();
