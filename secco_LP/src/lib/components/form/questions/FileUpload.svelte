<script lang="ts">
	import type { Question } from '$lib/config/questions';

	let { question, value = [], onchange }: {
		question: Question;
		value: File[];
		onchange: (files: File[]) => void;
	} = $props();

	let dragOver = $state(false);
	let inputEl: HTMLInputElement | undefined = $state();
	let error = $state('');

	let maxSize = $derived(question.maxFileSize ?? 10 * 1024 * 1024);

	function handleFiles(files: FileList | null) {
		if (!files || files.length === 0) return;
		error = '';

		const file = files[0];
		if (file.size > maxSize) {
			error = `Plik jest zbyt duży. Maksymalny rozmiar: ${Math.round(maxSize / 1024 / 1024)}MB`;
			return;
		}

		onchange([file]);
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		handleFiles(e.dataTransfer?.files ?? null);
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragOver = true;
	}

	function removeFile() {
		onchange([]);
		if (inputEl) inputEl.value = '';
	}
</script>

<div class="question">
	<p class="question-label">
		<span class="question-number">{question.number}.</span>
		{question.text}
	</p>

	{#if value.length === 0}
		<div
			class="dropzone"
			class:drag-over={dragOver}
			role="button"
			tabindex="0"
			ondrop={handleDrop}
			ondragover={handleDragOver}
			ondragleave={() => dragOver = false}
			onclick={() => inputEl?.click()}
			onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') inputEl?.click(); }}
			aria-label="Przeciągnij plik lub kliknij aby wybrać"
		>
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
				<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
				<polyline points="17 8 12 3 7 8" />
				<line x1="12" y1="3" x2="12" y2="15" />
			</svg>
			<p>Przeciągnij plik tutaj lub <strong>kliknij aby wybrać</strong></p>
			<p class="hint">Maks. {Math.round(maxSize / 1024 / 1024)}MB</p>
		</div>
		<input
			bind:this={inputEl}
			type="file"
			class="file-input"
			onchange={(e) => handleFiles(e.currentTarget.files)}
			aria-hidden="true"
			tabindex="-1"
		/>
	{:else}
		<div class="file-preview">
			<span class="file-name">{value[0].name}</span>
			<span class="file-size">({(value[0].size / 1024).toFixed(0)} KB)</span>
			<button type="button" class="remove-btn" onclick={removeFile} aria-label="Usuń plik">
				&times;
			</button>
		</div>
	{/if}

	{#if error}
		<p class="error">{error}</p>
	{/if}
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

	.dropzone {
		border: 2px dashed var(--color-accent-secondary);
		border-radius: var(--radius-md);
		padding: var(--spacing-3xl) var(--spacing-xl);
		text-align: center;
		cursor: pointer;
		transition: border-color var(--transition-fast), background-color var(--transition-fast);
		color: var(--color-text-light);
	}

	.dropzone:hover,
	.dropzone.drag-over {
		border-color: var(--color-accent);
		background-color: rgba(46, 49, 146, 0.04);
	}

	.dropzone svg {
		margin: 0 auto var(--spacing-md);
		color: var(--color-accent-secondary);
	}

	.dropzone p {
		margin-bottom: var(--spacing-xs);
	}

	.hint {
		font-size: var(--font-size-xs);
		color: var(--color-accent-secondary);
	}

	.file-input {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
	}

	.file-preview {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-md) var(--spacing-lg);
		border: 2px solid var(--color-accent);
		border-radius: var(--radius-md);
		background-color: rgba(46, 49, 146, 0.04);
	}

	.file-name {
		font-weight: 600;
		color: var(--color-text);
	}

	.file-size {
		font-size: var(--font-size-sm);
		color: var(--color-text-light);
	}

	.remove-btn {
		margin-left: auto;
		background: none;
		border: none;
		font-size: var(--font-size-xl);
		color: var(--color-text-light);
		cursor: pointer;
		padding: 0 var(--spacing-xs);
	}

	.remove-btn:hover {
		color: var(--color-error);
	}

	.error {
		color: var(--color-error);
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-sm);
	}
</style>
