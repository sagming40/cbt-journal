<script lang="ts">
	// S-04 단계별 입력 칸 (DESIGN 4.3절 표). 작성 · 수정 화면이 같이 쓴다.
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import AfterSliders from '#lib/components/AfterSliders.svelte';
	import DistortionPicker from '#lib/components/DistortionPicker.svelte';
	import EmotionPicker, { type Picked } from '#lib/components/EmotionPicker.svelte';
	import FormError from '#lib/components/FormError.svelte';
	import RecordTextarea from '#lib/components/RecordTextarea.svelte';
	import type { RecordForm } from '#lib/record-form.js';
	import { fromLocalInput, toLocalInput } from '#lib/tz.js';

	type Emotion = components['schemas']['Emotion'];
	type Distortion = components['schemas']['Distortion'];

	let {
		form = $bindable(),
		step,
		emotions,
		distortions,
		fields = {}
	}: {
		form: RecordForm;
		step: number;
		emotions: Emotion[];
		distortions: Distortion[];
		fields?: Record<string, string>;
	} = $props();

	const tz = $derived(auth.user?.timezone ?? 'Asia/Seoul');

	// EmotionPicker는 강도 하나를 다룬다 → 2단계의 "전"으로 잇는다. 이미 매긴 "후"는 그대로 둔다.
	const before = {
		get value(): Picked[] {
			return form.emotions.map((e) => ({
				emotion_id: e.emotion_id,
				intensity: e.intensity_before ?? 50
			}));
		},
		set value(picked: Picked[]) {
			form.emotions = picked.map((p) => ({
				emotion_id: p.emotion_id,
				intensity_before: p.intensity,
				intensity_after:
					form.emotions.find((e) => e.emotion_id === p.emotion_id)?.intensity_after ?? null
			}));
		}
	};

	// datetime-local은 지우면 ''을 준다. 시각을 비울 수는 없으니 그때는 이전 값을 둔다.
	function setWhen(value: string) {
		if (value) form.occurred_at = fromLocalInput(value, tz).toISOString();
	}
</script>

{#if step === 1}
	<div class="space-y-6">
		<RecordTextarea
			id="situation"
			label="상황"
			bind:value={form.situation}
			maxlength={1000}
			error={fields.situation}
		/>
		<div>
			<label for="occurred-at" class="text-sm text-muted-foreground">일어난 시각</label>
			<input
				id="occurred-at"
				type="datetime-local"
				value={toLocalInput(form.occurred_at, tz)}
				onchange={(ev) => setWhen(ev.currentTarget.value)}
				aria-invalid={fields.occurred_at ? 'true' : undefined}
				class="mt-1 block h-11 rounded-md border border-input bg-card px-3 tabular-nums"
			/>
			<div class="mt-1"><FormError message={fields.occurred_at} /></div>
		</div>
	</div>
{:else if step === 2}
	<EmotionPicker {emotions} bind:value={before.value} />
	<div class="mt-2"><FormError message={fields.emotions} /></div>
{:else if step === 3}
	<RecordTextarea
		id="automatic-thought"
		label="자동 사고"
		bind:value={form.automatic_thought}
		maxlength={1000}
		error={fields.automatic_thought}
	/>
{:else if step === 4}
	<DistortionPicker {distortions} bind:value={form.distortion_ids} />
	<div class="mt-2"><FormError message={fields.distortion_ids} /></div>
{:else if step === 5}
	<RecordTextarea
		id="evidence-for"
		label="근거"
		bind:value={form.evidence_for}
		maxlength={2000}
		error={fields.evidence_for}
	/>
{:else if step === 6}
	<RecordTextarea
		id="evidence-against"
		label="반대 근거"
		bind:value={form.evidence_against}
		maxlength={2000}
		error={fields.evidence_against}
	/>
{:else if step === 7}
	<RecordTextarea
		id="alternative-thought"
		label="대안 사고"
		bind:value={form.alternative_thought}
		maxlength={1000}
		error={fields.alternative_thought}
	/>
{:else}
	<AfterSliders {emotions} bind:value={form.emotions} />
	<div class="mt-2"><FormError message={fields.emotions} /></div>
{/if}
