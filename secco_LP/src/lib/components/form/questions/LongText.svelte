<script lang="ts">
	import type { Question } from '$lib/config/questions';

	let { question, value = '', onchange }: {
		question: Question;
		value: string;
		onchange: (value: string) => void;
	} = $props();

	let textareaEl: HTMLTextAreaElement | undefined = $state();

	export function focus() {
		textareaEl?.focus();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
		}
	}
</script>

<div class="question">
	<label for={question.id} class="question-label">
		<span class="question-number">{question.number}.</span>
		{question.text}
	</label>
	<textarea
		bind:this={textareaEl}
		id={question.id}
		class="long-text"
		{value}
		oninput={(e) => onchange(e.currentTarget.value)}
		onkeydown={handleKeydown}
		placeholder="Wpisz odpowiedź..."
		rows={4}
	></textarea>
	<p class="hint">Shift + Enter dla nowej linii</p>
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

	.long-text {
		width: 100%;
		padding: var(--spacing-md);
		border: 2px solid var(--color-accent-secondary);
		border-radius: var(--radius-md);
		background: transparent;
		font-family: var(--font-family);
		font-size: var(--font-size-base);
		color: var(--color-text);
		outline: none;
		resize: vertical;
		min-height: 120px;
		transition: border-color var(--transition-fast);
	}

	.long-text:focus {
		border-color: var(--color-accent);
	}

	.long-text::placeholder {
		color: var(--color-accent-secondary);
		opacity: 0.6;
	}

	.hint {
		font-size: var(--font-size-xs);
		color: var(--color-accent-secondary);
		margin-top: var(--spacing-sm);
	}
</style>
