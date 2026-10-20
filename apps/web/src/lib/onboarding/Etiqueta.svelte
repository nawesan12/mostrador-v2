<script lang="ts">
	/** Rótulo con punto y título de cada paso (handoff). */
	let {
		rotulo,
		titulo,
		grande = false
	}: { rotulo: string; titulo: string; grande?: boolean } = $props();
</script>

<div class="rotulo"><span></span>{rotulo}</div>
<h1 class:grande>{titulo}</h1>

<style>
	.rotulo {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #d2582f;
	}
	.rotulo span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
	}
	h1 {
		margin: 12px 0 0;
		font-size: clamp(32px, 3.6vw, 46px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	h1.grande {
		font-size: clamp(34px, 4vw, 52px);
		line-height: 1;
		letter-spacing: -0.045em;
	}
</style>
