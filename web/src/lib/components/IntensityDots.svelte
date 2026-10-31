<script lang="ts">
	// 강도 점 5개 — 채운 개수 = 강도 ÷ 20 반올림 (DESIGN 6장 "스타일 D2", 5.2절).
	// 색은 감정마다 다르게 하지 않고 개수로만 말한다(좋고 나쁨을 말하지 않게, DESIGN 1장 원칙 2).
	// 점 색 --dot: 라이트 accent, 다크 accent-text(다크 바탕 대비 3:1을 넘기려고).
	// 숫자를 바로 옆에 같이 보여 주므로 점은 화면 낭독기에서 건너뛴다.
	let { value }: { value: number } = $props();

	const filled = $derived(Math.round(Math.min(100, Math.max(0, value)) / 20));
</script>

<span class="inline-flex gap-[3px] align-middle" aria-hidden="true">
	{#each [0, 1, 2, 3, 4] as i (i)}
		<span class={['size-[7px] rounded-full border border-dot', i < filled && 'bg-dot']}></span>
	{/each}
</span>
