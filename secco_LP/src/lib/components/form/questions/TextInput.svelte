<script lang="ts">
	import type { Question } from '$lib/config/questions';

	let { question, value = '', onchange }: {
		question: Question;
		value: string;
		onchange: (value: string) => void;
	} = $props();

	let inputEl: HTMLInputElement | undefined = $state();

	export function focus() {
		inputEl?.focus();
	}
</script>

<div class="question">
	<label for={question.id} class="question-label">
		<span class="question-number">{question.number}.</span>
		{question.text}
	</label>
	{#if question.subtitle}
		<p class="question-subtitle">{question.subtitle}</p>
	{/if}
	<input
		bind:this={inputEl}
		id={question.id}
		type="text"
		class="text-input"
		{value}
		oninput={(e) => onchange(e.currentTarget.value)}
		placeholder="Wpisz odpowiedź..."
		autocomplete="off"
	/>
</div>

<style>
	.question {
		width: 100%;
		max-width: 600px;
	}

	.question-label {
		display: block;
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--color-text);
		margin-bottom: var(--spacing-lg);
		line-height: 1.4;
	}

	.question-number {
		color: var(--color-accent);
		margin-right: var(--spacing-xs);
	}

	.question-subtitle {
		font-size: var(--font-size-sm);
		color: var(--color-text-light);
		margin-bottom: var(--spacing-lg);
	}

	.text-input {
		width: 100%;
		padding: var(--spacing-md);
		border: none;
		border-bottom: 2px solid var(--color-accent-secondary);
		background: transparent;
		font-family: var(--font-family);
		font-size: var(--font-size-lg);
		color: var(--color-text);
		outline: none;
		transition: border-color var(--transition-fast);
	}

	.text-input:focus {
		border-bottom-color: var(--color-accent);
	}

	.text-input::placeholder {
		color: var(--color-accent-secondary);
		opacity: 0.6;
	}
</style>
