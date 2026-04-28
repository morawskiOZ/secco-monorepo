<script lang="ts">
	let { current, total }: { current: number; total: number } = $props();
	let percentage = $derived(Math.round((current / total) * 100));
</script>

<div class="progress" role="progressbar" aria-valuenow={percentage} aria-valuemin={0} aria-valuemax={100} aria-label="Postęp formularza">
	<div class="progress-bar" style="width: {percentage}%"></div>
</div>

<style>
	.progress {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 4px;
		background-color: rgba(0, 0, 0, 0.08);
		z-index: 100;
	}

	.progress-bar {
		height: 100%;
		background-color: var(--color-accent);
		transition: width var(--transition-normal);
	}
</style>
