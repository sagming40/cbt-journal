<script lang="ts">
	// "앞에서 쓴 답" (DESIGN 4.3절): 지금 단계보다 앞의 답만 읽기 전용으로, 누르면 그 단계로 이동.
	// 데스크톱은 오른쪽 패널, 휴대폰은 본문 위 접힌 카드에 같은 내용을 넣는다 — 감싸는 쪽(StepShell)이 정한다.
	import type { components } from '#lib/api/schema.js';
	import type { RecordForm } from '#lib/record-form.js';

	type Named = components['schemas']['Emotion'];

	let {
		form,
		step,
		emotions,
		distortions,
		onjump
	}: {
		form: RecordForm;
		step: number;
		emotions: Named[];
		distortions: Named[];
		onjump: (step: number) => void;
	} = $props();

	const emotionText = $derived(
		form.emotions
			.map((e) => {
				const name = emotions.find((x) => x.id === e.emotion_id)?.name ?? '';
				return `${name} ${e.intensity_before ?? ''}`.trim();
			})
			.join(' · ')
	);
	const distortionText = $derived(
		form.distortion_ids
			.map((id) => distortions.find((x) => x.id === id)?.name ?? '')
			.filter(Boolean)
			.join(' · ')
	);
	const rows = $derived(
		[
			{ step: 1, label: '상황', text: form.situation },
			{ step: 2, label: '감정', text: emotionText },
			{ step: 3, label: '자동 사고', text: form.automatic_thought },
			{ step: 4, label: '고른 왜곡', text: distortionText },
			{ step: 5, label: '근거', text: form.evidence_for },
			{ step: 6, label: '반대 근거', text: form.evidence_against },
			{ step: 7, label: '대안 사고', text: form.alternative_thought }
		].filter((r) => r.step < step && r.text)
	);
</script>

<ul class="space-y-1">
	{#each rows as r (r.step)}
		<li>
			<button
				type="button"
				class="-mx-2 block w-[calc(100%+1rem)] rounded-md px-2 py-2 text-left hover:bg-muted"
				onclick={() => onjump(r.step)}
			>
				<span class="block text-xs text-muted-foreground">{r.label}</span>
				<span class="line-clamp-4 block text-sm leading-relaxed whitespace-pre-wrap">{r.text}</span>
			</button>
		</li>
	{/each}
</ul>
