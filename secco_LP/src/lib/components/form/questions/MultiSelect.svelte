<script lang="ts">
	import type { Question } from '$lib/config/questions';
	import { onMount } from 'svelte';

	let { question, value = [], otherValue = '', onchange, onotherchange }: {
		question: Question;
		value: string[];
		otherValue?: string;
		onchange: (value: string[]) => void;
		onotherchange?: (value: string) => void;
	} = $props();

	const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';

	let otherSelected = $derived(value.includes('other'));
	let otherInputEl: HTMLInputElement | undefined = $state();

	function toggle(optionValue: string) {
		if (value.includes(optionValue)) {
			onchange(value.filter((v) => v !== optionValue));
		} else {
			if (question.maxSelections && value.length >= question.maxSelections) return;
			onchange([...value, optionValue]);
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!question.options) return;
		const totalOptions = question.options.length + (question.allowOther ? 1 : 0);
		const idx = letters.indexOf(e.key.toUpperCase());
		if (idx >= 0 && idx < totalOptions) {
			if (idx < question.options.length) {
				toggle(question.options[idx].value);
			} else {
				toggle('other');
			}
		}
	}

	onMount(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});

	$effect(() => {
		if (otherSelected && otherInputEl) {
			otherInputEl.focus();
		}
	});
</script>

<div class="question">
	<p class="question-label">
		<span class="question-number">{question.number}.</span>
		{question.text}
	</p>
	<p class="hint">{question.maxSelections ? `Wybierz maksymalnie ${question.maxSelections} opcje` : 'Wybierz dowolną liczbę opcji'}</p>
	<div class="options" role="group" aria-label={question.text}>
		{#each question.options ?? [] as option, i}
			<button
				type="button"
				class="option"
				class:selected={value.includes(option.value)}
				onclick={() => toggle(option.value)}
				role="checkbox"
				aria-checked={value.includes(option.value)}
			>
				<span class="option-letter">{letters[i]}</span>
				<span class="option-label">{option.label}</span>
			</button>
		{/each}
		{#if question.allowOther}
			{@const otherIdx = question.options?.length ?? 0}
			<button
				type="button"
				class="option"
				class:selected={otherSelected}
				onclick={() => toggle('other')}
				role="checkbox"
				aria-checked={otherSelected}
			>
				<span class="option-letter">{letters[otherIdx]}</span>
				<span class="option-label">Inne</span>
			</button>
			{#if otherSelected}
				<input
					bind:this={otherInputEl}
					type="text"
					class="other-input"
					value={otherValue}
					oninput={(e) => onotherchange?.(e.currentTarget.value)}
					onkeydown={(e) => { if (e.key.length === 1) e.stopPropagation(); }}
					placeholder="Wpisz swój styl..."
				/>
			{/if}
		{/if}
	</div>
</div>

<style>
	.question {
		width: 100%;
		max-width: 600px;
	}

	.question-label {
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--color-text);
		margin-bottom: var(--spacing-sm);
		line-height: 1.4;
	}

	.question-number {
		color: var(--color-accent);
		margin-right: var(--spacing-xs);
	}

	.hint {
		font-size: var(--font-size-sm);
		color: var(--color-accent-secondary);
		margin-bottom: var(--spacing-lg);
	}

	.options {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-md);
	}

	.option {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
		padding: var(--spacing-md) var(--spacing-lg);
		border: 2px solid var(--color-accent-secondary);
		border-radius: var(--radius-md);
		background: transparent;
		cursor: pointer;
		font-family: var(--font-family);
		font-size: var(--font-size-base);
		color: var(--color-text);
		text-align: left;
		transition: border-color var(--transition-fast), background-color var(--transition-fast);
	}

	.option:hover {
		border-color: var(--color-accent);
	}

	.option.selected {
		border-color: var(--color-accent);
		background-color: rgba(46, 49, 146, 0.08);
	}

	.option-letter {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border: 1px solid var(--color-accent-secondary);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-weight: 600;
		flex-shrink: 0;
	}

	.option.selected .option-letter {
		background-color: var(--color-accent);
		color: var(--color-white);
		border-color: var(--color-accent);
	}

	.other-input {
		width: 100%;
		padding: var(--spacing-md);
		border: none;
		border-bottom: 2px solid var(--color-accent);
		background: transparent;
		font-family: var(--font-family);
		font-size: var(--font-size-base);
		color: var(--color-text);
		outline: none;
		margin-top: var(--spacing-xs);
	}

	.other-input::placeholder {
		color: var(--color-accent-secondary);
		opacity: 0.6;
	}
</style>
