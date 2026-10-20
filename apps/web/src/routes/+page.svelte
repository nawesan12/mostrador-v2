<svelte:head>
	<title>Mostrador · Tu tienda online</title>
	<meta name="description" content="Todo lo que tienen las otras. Por menos plata y más seguro." />
</svelte:head>

<main>
	<img src="/wordmark-brown.png" alt="Mostrador" width="220" />
	<h1>Todo lo que tienen las otras. Por menos plata y más seguro.</h1>
	<p>Estamos armando Mostrador. Muy pronto vas a poder crear tu tienda acá.</p>
</main>

<style>
	main {
		min-height: 100svh;
		box-sizing: border-box;
		display: grid;
		place-content: center;
		justify-items: center;
		gap: 20px;
		padding: 24px 16px;
		text-align: center;
	}
	img {
		width: 220px;
		height: auto;
	}
	h1 {
		margin: 0;
		max-width: 16ch;
		font-size: clamp(34px, 6vw, 64px);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.045em;
		text-wrap: balance;
	}
	p {
		margin: 0;
		max-width: 38ch;
		font-size: 17px;
		line-height: 1.5;
		color: #6b5244;
		text-wrap: pretty;
	}
</style>
