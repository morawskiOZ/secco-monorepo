<script lang="ts">
	import { fly } from 'svelte/transition';
	import { formConfig, type FormConfig, type FormAnswers } from '$lib/config/questions';
	import ProgressBar from './ProgressBar.svelte';
	import NavigationButtons from './NavigationButtons.svelte';
	import WelcomeScreen from './WelcomeScreen.svelte';
	import ThankYouScreen from './ThankYouScreen.svelte';
	import TurnstileWidget from './TurnstileWidget.svelte';
	import GdprNotice from './GdprNotice.svelte';
	import TextInput from './questions/TextInput.svelte';
	import LongText from './questions/LongText.svelte';
	import RadioSelect from './questions/RadioSelect.svelte';
	import MultiSelect from './questions/MultiSelect.svelte';
	import FileUpload from './questions/FileUpload.svelte';
	import ContactForm from './questions/ContactForm.svelte';

	let { config = formConfig, submitUrl = '/api/submit' }: { config?: FormConfig; submitUrl?: string } = $props();

	type Screen = 'welcome' | 'questions' | 'thankyou';

	let screen = $state<Screen>('welcome');
	let step = $state(0);
	let answers = $state<FormAnswers>({});
	let turnstileToken = $state('');
	let submitting = $state(false);
	let submitError = $state('');
	let direction = $state(1);

	let contactFormRef: ContactForm | undefined = $state();

	const questions = config.questions;
	const totalSteps = questions.length;
	const currentQuestion = $derived(questions[step]);
	const isLastStep = $derived(step === totalSteps - 1);

	function start() {
		screen = 'questions';
	}

	function next() {
		if (isLastStep) return;

		if (currentQuestion.type === 'contact_form' && contactFormRef) {
			if (!contactFormRef.validate()) return;
		}

		direction = 1;
		step++;
	}

	function back() {
		if (step <= 0) return;
		direction = -1;
		step--;
	}

	function updateAnswer(questionId: string, value: string | string[] | File[] | Record<string, string>) {
		if (typeof value === 'object' && !Array.isArray(value) && !(value instanceof File)) {
			// Contact form: spread fields into answers
			for (const [k, v] of Object.entries(value as Record<string, string>)) {
				answers[k] = v;
			}
		} else {
			answers[questionId] = value as string | string[] | File[];
		}
	}

	async function submit() {
		if (currentQuestion.type === 'contact_form' && contactFormRef) {
			if (!contactFormRef.validate()) return;
		}

		submitting = true;
		submitError = '';

		try {
			const formData = new FormData();

			for (const [key, val] of Object.entries(answers)) {
				if (Array.isArray(val)) {
					if (val.length > 0 && val[0] instanceof File) {
						for (const file of val as File[]) {
							formData.append(key, file);
						}
					} else {
						formData.append(key, (val as string[]).join(', '));
					}
				} else if (typeof val === 'string') {
					formData.append(key, val);
				}
			}

			formData.append('cf-turnstile-response', turnstileToken);

			const resp = await fetch(submitUrl, {
				method: 'POST',
				body: formData
			});

			const data = await resp.json();

			if (!resp.ok) {
				submitError = data.error || 'Wystąpił błąd. Spróbuj ponownie.';
				return;
			}

			screen = 'thankyou';
		} catch {
			submitError = 'Nie udało się wysłać formularza. Sprawdź połączenie i spróbuj ponownie.';
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (screen === 'welcome') {
			if (e.key === 'Enter') {
				e.preventDefault();
				start();
			}
			return;
		}

		if (screen !== 'questions') return;
		if (e.target instanceof HTMLTextAreaElement) return;
		if (e.target instanceof HTMLInputElement && e.target.type !== 'radio' && e.key !== 'Enter') return;

		if (e.key === 'Enter') {
			e.preventDefault();
			if (isLastStep) {
				submit();
			} else {
				next();
			}
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="form-container">
	{#if screen === 'welcome'}
		<WelcomeScreen {config} onstart={start} />
	{:else if screen === 'thankyou'}
		<ThankYouScreen {config} />
	{:else}
		<ProgressBar current={step + 1} total={totalSteps} />

		<div class="question-area">
			{#key step}
				<div
					class="question-slide"
					in:fly={{ x: direction * 80, duration: 300, delay: 150 }}
					out:fly={{ x: direction * -80, duration: 150 }}
				>
					{#if currentQuestion.type === 'text'}
						<TextInput
							question={currentQuestion}
							value={(answers[currentQuestion.id] as string) ?? ''}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
						/>
					{:else if currentQuestion.type === 'long_text'}
						<LongText
							question={currentQuestion}
							value={(answers[currentQuestion.id] as string) ?? ''}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
						/>
					{:else if currentQuestion.type === 'radio'}
						<RadioSelect
							question={currentQuestion}
							value={(answers[currentQuestion.id] as string) ?? ''}
							otherValue={typeof answers[currentQuestion.id + '_other'] === 'string' ? answers[currentQuestion.id + '_other'] as string : ''}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
							onotherchange={(v) => updateAnswer(currentQuestion.id + '_other', v)}
						/>
					{:else if currentQuestion.type === 'multi_select'}
						<MultiSelect
							question={currentQuestion}
							value={Array.isArray(answers[currentQuestion.id]) ? answers[currentQuestion.id] as string[] : []}
							otherValue={typeof answers[currentQuestion.id + '_other'] === 'string' ? answers[currentQuestion.id + '_other'] as string : ''}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
							onotherchange={(v) => updateAnswer(currentQuestion.id + '_other', v)}
						/>
					{:else if currentQuestion.type === 'file_upload'}
						<FileUpload
							question={currentQuestion}
							value={Array.isArray(answers[currentQuestion.id]) ? answers[currentQuestion.id] as File[] : []}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
						/>
					{:else if currentQuestion.type === 'contact_form'}
						<ContactForm
							bind:this={contactFormRef}
							question={currentQuestion}
							value={{
								first_name: typeof answers.first_name === 'string' ? answers.first_name : '',
								last_name: typeof answers.last_name === 'string' ? answers.last_name : '',
								phone: typeof answers.phone === 'string' ? answers.phone : '',
								email: typeof answers.email === 'string' ? answers.email : ''
							}}
							onchange={(v) => updateAnswer(currentQuestion.id, v)}
						/>
					{/if}

					{#if isLastStep}
						<GdprNotice {config} />
						<TurnstileWidget ontoken={(t) => turnstileToken = t} />
					{/if}

					{#if submitError}
						<p class="submit-error">{submitError}</p>
					{/if}

					<NavigationButtons
						showBack={step > 0}
						showNext={!isLastStep}
						showSubmit={isLastStep}
						{submitting}
						onback={back}
						onnext={next}
						onsubmit={submit}
					/>
				</div>
			{/key}
		</div>
	{/if}
</div>

<style>
	.form-container {
		min-height: 100vh;
		background-color: var(--color-bg-page);
		display: flex;
		flex-direction: column;
	}

	.question-area {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-3xl) var(--spacing-xl);
		position: relative;
		overflow: hidden;
	}

	.question-slide {
		width: 100%;
		max-width: 640px;
	}

	.submit-error {
		color: var(--color-error);
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-md);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: rgba(211, 47, 47, 0.08);
		border-radius: var(--radius-sm);
	}
</style>
