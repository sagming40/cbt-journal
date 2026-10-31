<script lang="ts">
	// 감정 고르기 — 칩 + 고른 감정마다 강도 슬라이더 (DESIGN 4.2절, 5.2절 EmotionPicker).
	// 칩을 고르면 강도는 50에서 시작한다(0에서 시작하면 매번 끌어야 한다). 슬라이더는 브라우저 기본 range라
	// 화살표 5 · PageUp/Down 10으로 조절된다 (DESIGN 8장 접근성).
	import type { components } from '#lib/api/schema.js';

	type Emotion = components['schemas']['Emotion'];
	export type Picked = { emotion_id: number; intensity: number };

	// 오류 문구는 감싸는 폼의 FormError(role=alert)가 읽어 준다 — fieldset에는 aria-invalid를 달 수 없어서
	let {
		emotions,
		value = $bindable([])
	}: {
		emotions: Emotion[];
		value?: Picked[];
	} = $props();

	const START = 50;
	const picked = (id: number) => value.some((p) => p.emotion_id === id);
	// 슬라이더는 감정 화면 순서대로 — 고른 순서와 상관없이 늘 같은 자리에 있게(서버 응답 순서와도 같다)
	const rows = $derived(
		emotions
			.map((e) => ({ emotion: e, pick: value.find((p) => p.emotion_id === e.id) }))
			.filter((r) => r.pick)
	);

	function toggle(id: number) {
		value = picked(id)
			? value.filter((p) => p.emotion_id !== id)
			: [...value, { emotion_id: id, intensity: START }];
	}

	function setIntensity(id: number, intensity: number) {
		value = value.map((p) => (p.emotion_id === id ? { ...p, intensity } : p));
	}
</script>

<fieldset>
	<legend class="sr-only">감정</legend>
	<div class="flex flex-wrap gap-2">
		{#each emotions as e (e.id)}
			<button
				type="button"
				aria-pressed={picked(e.id)}
				onclick={() => toggle(e.id)}
				class={[
					'h-11 rounded-md border px-4 text-[15px] transition-colors',
					picked(e.id)
						? 'border-primary bg-accent-tint font-bold text-accent-text shadow-[inset_0_-3px_0_var(--primary)] dark:border-accent-text dark:shadow-[inset_0_-3px_0_var(--accent-text)]'
						: 'border-input bg-card hover:bg-muted'
				]}>{e.name}</button
			>
		{/each}
	</div>
</fieldset>

{#if rows.length > 0}
	<ul class="mt-4">
		{#each rows as { emotion, pick } (emotion.id)}
			<li
				class="grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1 border-b border-dashed py-3.5"
			>
				<label for={`intensity-${emotion.id}`} class="font-serif text-[17px] font-bold"
					>{emotion.name}</label
				>
				<span class="text-xl font-bold tabular-nums" aria-hidden="true">{pick!.intensity}</span>
				<input
					id={`intensity-${emotion.id}`}
					type="range"
					min="0"
					max="100"
					step="5"
					value={pick!.intensity}
					oninput={(ev) => setIntensity(emotion.id, Number(ev.currentTarget.value))}
					aria-label={`${emotion.name} 강도`}
					aria-valuetext={`${pick!.intensity}`}
					class="col-span-2 h-6 w-full accent-primary dark:accent-accent-text"
				/>
			</li>
		{/each}
	</ul>
{/if}
