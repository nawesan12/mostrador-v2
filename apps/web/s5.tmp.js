await page.goto(BASE + '/onboarding?paso=1', { waitUntil: 'networkidle' });
console.log('on reload:', (await txt('body')).slice(0, 120));
for (let step = 0; step < 8; step++) {
  const t = await txt('body');
  const m = t.match(/Paso (\d) de 8/); const n = m ? m[1] : 'x';
  await shot('m_onb_step' + n);
  console.log('\n### STEP', n, page.url()); console.log(t.slice(0, 1400));
  console.log(JSON.stringify(await ctrls()).slice(0, 1500));
  const of = await overflow(); if (of.sw > of.vw) console.log('OVERFLOW', JSON.stringify(of));
  if (n === '3') { await page.getByRole('button', { name: 'Continuar' }).click(); }
  else break;
  await sleep(2500);
}
