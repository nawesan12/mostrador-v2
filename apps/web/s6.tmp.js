await page.goto(BASE + '/onboarding?paso=1', { waitUntil: 'networkidle' });
console.log('pressed after reload:', await page.evaluate(() => [...document.querySelectorAll('button[aria-pressed=true]')].map(b => b.innerText)));
for (let i = 0; i < 12; i++) {
  const t = await txt('body');
  const m = t.match(/Paso (\d) de 8/); const n = m ? m[1] : 'x';
  await shot('m_onb_step' + n + '_' + i);
  console.log('\n### STEP', n, page.url()); console.log(t.slice(0, 1600));
  console.log(JSON.stringify(await ctrls()).slice(0, 1800));
  const of = await overflow(); if (of.sw > of.vw) console.log('OVERFLOW', JSON.stringify(of));
  if (n === '2') { const b = page.getByRole('button', { name: 'Vender online por primera vez' }); if ((await b.getAttribute('aria-pressed')) !== 'true') await b.click(); }
  if (n === '4' || n === '5' || n === '6' || n === '7' || n === '8' || n === 'x') break;
  await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(2500);
}
