<script lang="ts">
	// 로고 · 비켜 선 두 원 (DESIGN 6장 "로고 · 부제").
	// 찬 원 = 지금 마음, 한 걸음 떨어진 빈 원 = 떨어져서 다시 본 마음. 둘 사이의 틈이 "한 걸음".
	// 효과(유리 · 그라데이션 · 3D)는 입히지 않는다 — 납작한 단색만. 색은 화면 안 로고 규칙: 라이트 accent, 다크 accent-text.
	let {
		size = 32,
		wordmark = false,
		tagline = false,
		class: className = ''
	}: {
		/** 상징 크기(px) */
		size?: number;
		/** 옆에 "CBT 저널" 글자를 붙인다 */
		wordmark?: boolean;
		/** 글자 아래 부제를 붙인다 (wordmark일 때만) */
		tagline?: boolean;
		class?: string;
	} = $props();
</script>

<span class={['inline-flex items-center gap-2.5', className]}>
	<!-- 모양 수치는 DESIGN 6장: 64 칸 안에서 (22, 36) r14 찬 원 / (43, 26) r14 빈 원, 선 3.5 -->
	<svg
		width={size}
		height={size}
		viewBox="0 0 64 64"
		class="shrink-0 text-primary dark:text-accent-text"
		aria-hidden={wordmark ? 'true' : undefined}
		role={wordmark ? undefined : 'img'}
		aria-label={wordmark ? undefined : 'CBT 저널'}
	>
		<circle cx="22" cy="36" r="14" fill="currentColor" />
		<circle cx="43" cy="26" r="14" fill="none" stroke="currentColor" stroke-width="3.5" />
	</svg>
	{#if wordmark}
		<span class="flex flex-col">
			<span class="font-serif text-lg leading-tight font-bold">CBT 저널</span>
			{#if tagline}
				<span class="text-xs text-muted-foreground">한 걸음 떨어져, 내 마음을 다시 보기</span>
			{/if}
		</span>
	{/if}
</span>
