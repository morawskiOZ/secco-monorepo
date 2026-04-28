<footer class="footer">
	<div class="footer-content">
		<div class="footer-social">
			<a
				href="https://www.instagram.com/secco.studio/"
				target="_blank"
				rel="noopener noreferrer"
				aria-label="Instagram Secco Studio"
				class="instagram-link"
			>
				<svg class="instagram-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
					<rect x="2" y="2" width="20" height="20" rx="5" stroke="currentColor" stroke-width="1.5"/>
					<circle cx="12" cy="12" r="5" stroke="currentColor" stroke-width="1.5"/>
					<circle cx="17.5" cy="6.5" r="1" fill="currentColor"/>
				</svg>
				secco.studio
			</a>
		</div>
		<div class="footer-keywords" aria-hidden="true">
			<p>architektura · wnętrza · architekt wnętrz · wrocław · legnica · projekt · kuchnia · projektowanie wnętrz · wizualizacje</p>
		</div>
	</div>
</footer>

<style>
	.footer {
		background-color: var(--color-bg-primary);
		padding: var(--spacing-lg) var(--spacing-2xl) var(--spacing-3xl);
	}

	.footer-content {
		max-width: var(--max-width);
		margin: 0 auto;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--spacing-md);
	}

	.footer-social {
		display: flex;
		align-items: center;
	}

	.instagram-link {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		font-size: var(--font-size-sm);
		color: var(--color-text);
		text-decoration: none;
		transition: color var(--transition-fast);
	}

	.instagram-link:hover {
		color: var(--color-accent);
	}

	.instagram-icon {
		width: 40px;
		height: 40px;
		flex-shrink: 0;
	}

	.footer-keywords {
		flex: 1;
	}

	.footer-keywords p {
		font-size: var(--font-size-xs);
		color: var(--color-accent-secondary);
		text-align: center;
	}

	@media (max-width: 768px) {
		.footer {
			padding: var(--spacing-md) var(--spacing-lg);
		}

		.footer-content {
			flex-direction: column;
			text-align: center;
			gap: var(--spacing-sm);
		}

		.footer-keywords p {
			text-align: center;
		}
	}
</style>
