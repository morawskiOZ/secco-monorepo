<script>
	import Hero from '$lib/components/landing/Hero.svelte';
	import Footer from '$lib/components/landing/Footer.svelte';
</script>

<svelte:head>
	<title>Secco Studio — Projektowanie wnętrz Wrocław</title>
	<meta name="description" content="Secco Studio — studio projektowe zlokalizowane we Wrocławiu. Kompleksowa aranżacja wnętrz domów i mieszkań. Projektujemy, doradzamy, nadzorujemy." />
	<meta name="keywords" content="architektura, wnętrza, architekt wnętrz, wrocław, legnica, projekt, kuchnia, projektowanie wnętrz, wizualizacje" />
	<link rel="canonical" href="https://www.seccostudio.com/" />

	<!-- Open Graph -->
	<meta property="og:type" content="website" />
	<meta property="og:title" content="Secco Studio — Projektowanie wnętrz Wrocław" />
	<meta property="og:description" content="Kompleksowa aranżacja wnętrz domów i mieszkań we Wrocławiu." />
	<meta property="og:url" content="https://www.seccostudio.com/" />
	<meta property="og:site_name" content="Secco Studio" />
	<meta property="og:locale" content="pl_PL" />
	<meta property="og:image" content="https://www.seccostudio.com/images/hero.webp" />

	<!-- JSON-LD -->
	{@html `<script type="application/ld+json">${JSON.stringify({
		"@context": "https://schema.org",
		"@type": "InteriorDesigner",
		"name": "Secco Studio",
		"url": "https://www.seccostudio.com",
		"telephone": "+48665895432",
		"email": "secco.studio@gmail.com",
		"address": {
			"@type": "PostalAddress",
			"addressLocality": "Wrocław",
			"addressCountry": "PL"
		},
		"description": "Studio projektowe zlokalizowane we Wrocławiu. Kompleksowa aranżacja wnętrz domów i mieszkań.",
		"sameAs": [
			"https://www.instagram.com/secco.studio/",
			"https://issuu.com/secco.studio"
		]
	})}</script>`}
</svelte:head>

<div class="page">
	<Hero />
	<Footer />
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}
</style>
