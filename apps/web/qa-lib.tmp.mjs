import { chromium } from '@playwright/test';
export const SS = '/private/tmp/claude-501/-Users-nahuelsantillan-Trabajo-vimo/0b2cacf3-b3fd-4f58-b5f5-059480d242be/scratchpad/qa/';
export const BASE = 'https://tulocalweb.com.ar';
export async function ctx(browser, mobile, storage) {
  const c = await browser.newContext({
    viewport: mobile ? { width: 375, height: 812 } : { width: 1440, height: 900 },
    isMobile: !!mobile, hasTouch: !!mobile, locale: 'es-AR',
    storageState: storage, acceptDownloads: true,
    userAgent: mobile ? 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1' : undefined,
  });
  return c;
}
export function watch(page, log) {
  page.on('console', m => { if (m.type() === 'error') log.push(`[console] ${page.url()} :: ${m.text().slice(0,300)}`); });
  page.on('pageerror', e => log.push(`[pageerror] ${page.url()} :: ${String(e).slice(0,300)}`));
  page.on('response', r => { if (r.status() >= 400) log.push(`[http ${r.status()}] ${r.request().method()} ${r.url()} (from ${page.url()})`); });
  page.on('requestfailed', r => { const f = r.failure()?.errorText||''; if (!/ERR_ABORTED/.test(f)) log.push(`[reqfail] ${r.url()} ${f}`); });
}
export async function overflow(page) {
  return page.evaluate(() => {
    const vw = window.innerWidth; const sw = document.documentElement.scrollWidth;
    const offenders = [];
    if (sw > vw + 1) for (const el of document.querySelectorAll('body *')) {
      const r = el.getBoundingClientRect(); if (r.right > vw + 1 && r.width > 0 && getComputedStyle(el).position !== 'fixed') {
        offenders.push((el.tagName + '.' + (el.className?.baseVal ?? el.className)).slice(0,80) + ' r=' + Math.round(r.right)); if (offenders.length > 5) break; }
    }
    return { vw, sw, offenders };
  });
}
export async function cookies(page) {
  for (const t of ['Aceptar', 'Aceptar todo', 'Aceptar todas', 'Entendido', 'De acuerdo']) {
    const b = page.getByRole('button', { name: t, exact: true });
    if (await b.count()) { try { await b.first().click({ timeout: 1500 }); return t; } catch {} }
  }
  return null;
}
export { chromium };
