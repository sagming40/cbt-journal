<script lang="ts">
	// 8단계: 감정 다시 매기기 (DESIGN 4.3절). 2단계 감정마다 슬라이더, 전 강도는 흐리게 같이 보여 준다.
	// 줄었는지 늘었는지 색으로 말하지 않는다 — 숫자만 (1장 원칙 2). 이 화면이 곧 완료 전 요약이다.
	import type { components } from '#lib/api/schema.js';
	import IntensityChange from '#lib/components/IntensityChange.svelte';

	type Emotion = components['schemas']['Emotion'];
	type Item = components['schemas']['RecordEmotionInput'];

	let { emotions, value = $bindable([]) }: { emotions: Emotion[]; value?: Item[] } = $props();

	// 감정 화면 순서대로 — 2단계 · 서버 응답과 같은 자리에
	const rows = $derived(
		emotions
			.map((e) => ({ emotion: e, item: value.find((v) => v.emotion_id === e.id) }))
			.filter((r) => r.item)
	);

	function setAfter(id: number, after: number) {
		value = value.map((v) => (v.emotion_id === id ? { ...v, intensity_after: after } : v));
	}
</script>

<ul>
	{#each rows as { emotion, item } (emotion.id)}
		{@const before = item!.intensity_before ?? 0}
		{@const after = item!.intensity_after ?? before}
		<li class="border-b border-dashed py-4">
			<div class="flex items-baseline justify-between gap-3">
				<label for={`after-${emotion.id}`} class="font-serif text-[17px] font-bold"
					>{emotion.name}</label
				>
				<span class="text-xl font-bold"><IntensityChange {before} {after} /></span>
			</div>
			<div class="relative mt-2">
				<input
					id={`after-${emotion.id}`}
					type="range"
					min="0"
					max="100"
					step="5"
					value={after}
					oninput={(ev) => setAfter(emotion.id, Number(ev.currentTarget.value))}
					aria-label={`${emotion.name} 지금 강도`}
					aria-valuetext={`${after}, 전에는 ${before}`}
					class="h-6 w-full accent-primary dark:accent-accent-text"
				/>
				<!-- "전" 자리를 흐린 눈금으로. 손잡이 폭(약 16px) 때문에 양 끝이 안쪽으로 조금 들어간다 -->
				<span
					class="pointer-events-none absolute top-6 text-[11px] leading-none text-muted-foreground tabular-nums"
					style:left={`calc(${before}% + ${8 - before * 0.16}px)`}
					style:transform="translateX(-50%)"
					aria-hidden="true">▲전</span
				>
			</div>
			<div class="h-4"></div>
		</li>
	{/each}
</ul>
