<script lang="ts">
	import type { Question } from '$lib/config/questions';

	let { question, value = {}, onchange }: {
		question: Question;
		value: Record<string, string>;
		onchange: (value: Record<string, string>) => void;
	} = $props();

	let errors = $state<Record<string, string>>({});
	let inputEls: HTMLInputElement[] = $state([]);

	export function focus() {
		inputEls[0]?.focus();
	}

	export function validate(): boolean {
		errors = {};
		let valid = true;

		for (const field of question.fields ?? []) {
			const val = (value[field.id] ?? '').trim();
			if (field.required && !val) {
				errors[field.id] = 'To pole jest wymagane';
				valid = false;
			}
			if (field.type === 'email' && val && (!val.includes('@') || !val.includes('.'))) {
				errors[field.id] = 'Nieprawidłowy adres email';
				valid = false;
			}
		}

		return valid;
	}

	function updateField(fieldId: string, fieldValue: string) {
		onchange({ ...value, [fieldId]: fieldValue });
		if (errors[fieldId]) {
			errors = { ...errors, [fieldId]: '' };
		}
	}
</script>

<div class="question">
	<p class="question-label">
		<span class="question-number">{question.number}.</span>
		{question.text}
	</p>
	<div class="fields">
		{#each question.fields ?? [] as field, i}
			<div class="field">
				<label for={field.id} class="field-label">
					{field.label}
					{#if field.required}<span class="required">*</span>{/if}
				</label>
				<input
					bind:this={inputEls[i]}
					id={field.id}
					type={field.type}
					class="field-input"
					class:field-error={errors[field.id]}
					value={value[field.id] ?? ''}
					oninput={(e) => updateField(field.id, e.currentTarget.value)}
					placeholder={field.placeholder}
					autocomplete={field.type === 'email' ? 'email' : field.type === 'tel' ? 'tel' : 'off'}
				/>
				{#if errors[field.id]}
					<p class="error">{errors[field.id]}</p>
				{/if}
			</div>
		{/each}
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
		margin-bottom: var(--spacing-lg);
		line-height: 1.4;
	}

	.question-number {
		color: var(--color-accent);
		margin-right: var(--spacing-xs);
	}

	.fields {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-lg);
	}

	.field-label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--color-text);
		margin-bottom: var(--spacing-xs);
	}

	.required {
		color: var(--color-error);
	}

	.field-input {
		width: 100%;
		padding: var(--spacing-md);
		border: none;
		border-bottom: 2px solid var(--color-accent-secondary);
		background: transparent;
		font-family: var(--font-family);
		font-size: var(--font-size-base);
		color: var(--color-text);
		outline: none;
		transition: border-color var(--transition-fast);
	}

	.field-input:focus {
		border-bottom-color: var(--color-accent);
	}

	.field-input.field-error {
		border-bottom-color: var(--color-error);
	}

	.field-input::placeholder {
		color: var(--color-accent-secondary);
		opacity: 0.6;
	}

	.error {
		color: var(--color-error);
		font-size: var(--font-size-xs);
		margin-top: var(--spacing-xs);
	}
</style>
