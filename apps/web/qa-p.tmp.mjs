// persistent-session runner: bun qa-p.tmp.mjs <m|d> <script.js>  (script body gets page, ctx helpers)
import { chromium, watch, overflow, cookies, SS, BASE } from './qa-lib.tmp.mjs';
import fs from 'fs';
const [mode, file] = process.argv.slice(2);
const mobile = mode === 'm';
const dir = '/private/tmp/claude-501/-Users-nahuelsantillan-Trabajo-vimo/0b2cacf3-b3fd-4f58-b5f5-059480d242be/scratchpad/profile-merchant';
const context = await chromium.launchPersistentContext(dir, {
  viewport: mobile ? { width: 375, height: 812 } : { width: 1440, height: 900 }, isMobile: mobile, hasTouch: mobile, locale: 'es-AR', acceptDownloads: true,
  userAgent: mobile ? 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1' : undefined,
});
const page = context.pages()[0] || await context.newPage();
const log = []; watch(page, log);
const txt = async (sel = 'body') => (await page.evaluate(s => document.querySelector(s)?.innerText || '', sel)).replace(/\n+/g, ' | ');
const shot = async (n, full = true) => page.screenshot({ path: SS + n + '.png', fullPage: full });
const ctrls = async () => page.evaluate(() => [...document.querySelectorAll('input:not([type=hidden]),button,select,textarea,a[href]')].filter(e => e.offsetParent !== null).map(e => `${e.tagName}${e.name ? ':' + e.name : ''}${e.type && e.tagName==='INPUT' ? ':' + e.type : ''}:${(e.innerText || e.placeholder || e.getAttribute('aria-label') || e.value || '').trim().replace(/\n+/g,' ').slice(0, 40)}${e.tagName==='A' ? '->' + e.getAttribute('href') : ''}`));
const sleep = ms => page.waitForTimeout(ms);
const mail = (n = 3) => { const { execSync } = require('child_process'); return execSync(`ssh root@187.41.102.190 'python3 /root/qa_mail.py ${n}'`).toString(); };
const body = fs.readFileSync(file, 'utf8');
const fn = new (Object.getPrototypeOf(async function () {}).constructor)('page', 'context', 'txt', 'shot', 'ctrls', 'sleep', 'overflow', 'cookies', 'BASE', 'SS', 'mail', 'log', 'fs', body);
try { await fn(page, context, txt, shot, ctrls, sleep, () => overflow(page), () => cookies(page), BASE, SS, mail, log, fs); }
catch (e) { console.log('SCRIPT ERROR', e.message.slice(0, 500)); await shot('error_' + Date.now(), false).catch(()=>{}); }
console.log('URL at end:', page.url());
console.log('--- LOG'); console.log([...new Set(log)].join('\n'));
await context.close();
