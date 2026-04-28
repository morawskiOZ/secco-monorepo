<section class="hero" aria-label="Strona główna">
	<div class="hero-content">
		<div class="hero-text">
			<h1 class="hero-title">secco.studio</h1>
			<h2 class="hero-subtitle">Studio projektowe zlokalizowane we Wrocławiu</h2>
			<p class="hero-description">
				Zajmujemy się kompleksową aranżacją wnętrz domów i mieszkań.
				Projektujemy, doradzamy, nadzorujemy i staramy się sprawnie przeprowadzić
				Państwa przez cały proces projektowy.
			</p>
			<a href="/form" class="hero-cta">Wypełnij formularz wyceny</a>
		</div>
		<div class="hero-image-wrapper">
			<img
				src="/images/hero.webp"
				alt="Wnętrze kuchni zaprojektowane przez Secco Studio"
				width="800"
				height="1202"
				loading="eager"
				fetchpriority="high"
			/>
		</div>
	</div>
</section>

<style>
	.hero {
		background-color: var(--color-bg-primary);
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		overflow: hidden;
		padding: var(--spacing-3xl) var(--spacing-2xl);
	}

	.hero::before {
		content: '';
		position: absolute;
		top: -80px;
		right: -80px;
		width: 280px;
		height: 280px;
		border-radius: var(--radius-full);
		background-color: var(--color-accent);
		opacity: 0.9;
	}

	.hero::after {
		content: '';
		position: absolute;
		bottom: -60px;
		left: 50%;
		transform: translateX(-50%);
		width: 200px;
		height: 200px;
		border-radius: var(--radius-full);
		background-color: var(--color-accent-secondary);
		opacity: 0.6;
	}

	.hero-content {
		max-width: var(--max-width);
		width: 100%;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--spacing-3xl);
		align-items: center;
		position: relative;
		z-index: 1;
	}

	.hero-title {
		font-size: var(--font-size-3xl);
		font-weight: 700;
		color: var(--color-text);
		margin-bottom: var(--spacing-md);
		letter-spacing: 0.02em;
	}

	.hero-subtitle {
		font-size: var(--font-size-lg);
		font-weight: 400;
		color: var(--color-text-light);
		margin-bottom: var(--spacing-xl);
	}

	.hero-description {
		font-size: var(--font-size-base);
		color: var(--color-text-light);
		line-height: 1.8;
		margin-bottom: var(--spacing-2xl);
		max-width: 480px;
	}

	.hero-cta {
		display: inline-block;
		background-color: var(--color-accent);
		color: var(--color-white);
		padding: var(--spacing-md) var(--spacing-2xl);
		border-radius: var(--radius-md);
		font-size: var(--font-size-base);
		font-weight: 600;
		transition: opacity var(--transition-fast);
		text-decoration: none;
	}

	.hero-cta:hover {
		opacity: 0.9;
		text-decoration: none;
	}

	.hero-image-wrapper {
		border-radius: var(--radius-md);
		overflow: hidden;
		box-shadow: var(--shadow-md);
		max-height: 75vh;
	}

	.hero-image-wrapper img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	@media (max-width: 768px) {
		.hero {
			padding: var(--spacing-2xl) var(--spacing-lg);
		}

		.hero-content {
			grid-template-columns: 1fr;
			gap: var(--spacing-2xl);
			text-align: center;
		}

		.hero-title {
			font-size: var(--font-size-2xl);
		}

		.hero-description {
			max-width: 100%;
		}

		.hero-cta {
			width: 100%;
			text-align: center;
		}

		.hero::before {
			width: 160px;
			height: 160px;
			top: -40px;
			right: -40px;
		}

		.hero::after {
			width: 120px;
			height: 120px;
			bottom: -30px;
		}
	}

	@media (max-width: 500px) {
		.hero::before {
			width: 140px;
			height: 140px;
			top: -60px;
			right: -60px;
		}
	}
</style>
