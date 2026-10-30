await page.goto(BASE + '/onboarding', { waitUntil: 'networkidle' }); await cookies();
await page.fill('input[name=email]', 'nahuel+qa1@grupovimo.com.ar');
await page.getByRole('button', { name: 'Seguir' }).click();
await sleep(3000);
await shot('m_onb_1');
console.log(await txt('main'));
console.log(await ctrls());
