await page.goto(BASE + '/onboarding?paso=1', { waitUntil: 'networkidle' });
console.log('at', (await txt('body')).slice(0, 40));
await page.getByRole('button', { name: 'Ver las 17 plantillas' }).click(); await sleep(1200);
await shot('m_onb_17plantillas', false);
console.log('17:', (await txt('body')).slice(0, 900));
await page.keyboard.press('Escape'); await sleep(500);
await page.getByRole('button', { name: 'Yerbal' }).click(); await sleep(1000);
await page.getByRole('button', { name: 'Celular' }).click(); await sleep(800);
await shot('m_onb_yerbal_cel', false);
await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(3000);
for (let i = 0; i < 4; i++) {
  const t = await txt('body'); const n = (t.match(/Paso (\d) de 8/) || [])[1];
  console.log('\n### step', n, page.url()); console.log(t.slice(0, 2000));
  console.log(JSON.stringify(await ctrls()).slice(0, 2500));
  await shot('m_onb_step' + n + '_b');
  const of = await overflow(); if (of.sw > of.vw) console.log('OVERFLOW', JSON.stringify(of));
  if (n === '6') break;
  break;
}
