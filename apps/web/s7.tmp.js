await page.goto(BASE + '/onboarding?paso=1', { waitUntil: 'networkidle' });
// verify goal persistence: go to step 2 via stepper
await page.getByRole('button', { name: 'Tu negocio' }).first().click(); await sleep(800);
console.log('now', (await txt('body')).slice(0, 60));
for (let k=0;k<8;k++){ const b=page.locator('button[aria-pressed=true]'); if(!(await b.count())) break; await b.first().click(); await sleep(300);}
await page.getByRole('button', { name: 'Facturar sin complicarme' }).click();
console.log('pressed before', await page.evaluate(() => [...document.querySelectorAll('button[aria-pressed=true]')].map(b => b.innerText)));
await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(2500);
await page.reload({ waitUntil: 'networkidle' });
await page.getByRole('button', { name: 'Tu negocio' }).first().click(); await sleep(800);
console.log('pressed after reload', await page.evaluate(() => [...document.querySelectorAll('button[aria-pressed=true]')].map(b => b.innerText)));
await page.getByRole('button', { name: 'Vender online por primera vez' }).click();
await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(2000);
await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(2000);
console.log('at', (await txt('body')).slice(0, 60));
await page.getByRole('button', { name: 'Empiezo de cero' }).click(); await sleep(500);
await page.getByRole('button', { name: 'Continuar' }).click(); await sleep(3000);
for (let i = 0; i < 2; i++) {
  const t = await txt('body'); console.log('\n### ', page.url()); console.log(t.slice(0, 2500));
  console.log(JSON.stringify(await ctrls()).slice(0, 2500));
  await shot('m_onb_s5_' + i);
  const of = await overflow(); if (of.sw > of.vw) console.log('OVERFLOW', JSON.stringify(of));
  break;
}
