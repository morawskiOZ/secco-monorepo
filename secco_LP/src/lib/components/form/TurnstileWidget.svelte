<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	let { ontoken }: {
		ontoken: (token: string) => void;
	} = $props();

	let container: HTMLDivElement | undefined = $state();
	let widgetId: string | undefined = $state();

	import { PUBLIC_TURNSTILE_SITE_KEY } from '$env/static/public';

	const siteKey = PUBLIC_TURNSTILE_SITE_KEY;

	onMount(() => {
		const script = document.createElement('script');
		script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
		script.async = true;
		script.onload = () => {
			if (!container) return;
			widgetId = (window as any).turnstile.render(container, {
				sitekey: siteKey,
				theme: 'light',
				size: 'flexible',
				callback: (token: string) => ontoken(token),
				'expired-callback': () => {
					ontoken('');
					if (widgetId !== undefined) {
						(window as any).turnstile.reset(widgetId);
					}
				}
			});
		};
		document.head.appendChild(script);
	});

	onDestroy(() => {
		if (widgetId !== undefined && typeof (window as any).turnstile !== 'undefined') {
			(window as any).turnstile.remove(widgetId);
		}
	});
</script>

<div class="turnstile-wrapper">
	<div bind:this={container}></div>
</div>

<style>
	.turnstile-wrapper {
		margin-top: var(--spacing-md);
	}
</style>
