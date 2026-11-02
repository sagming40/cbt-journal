<script lang="ts">
	// 머리의 저장 상태 (DESIGN 4.3절 "자동 저장" · "저장 실패", 5.2절).
	// 저장되는지 늘 보이게 해서 "날아가면 어쩌지"를 덜어 준다. 실패는 빨강이 아니라 아이콘 + 글자 (6장).
	import CircleAlert from '@lucide/svelte/icons/circle-alert';
	import type { SaveState } from '#lib/draft-saver.js';
	import { formatAgo } from '#lib/tz.js';

	let {
		status,
		savedAt,
		onretry
	}: { status: SaveState; savedAt: Date | null; onretry: () => void } = $props();

	// "방금"이 계속 방금이지 않게 30초마다 다시 계산한다
	let now = $state(new Date());
	$effect(() => {
		const id = setInterval(() => (now = new Date()), 30_000);
		return () => clearInterval(id);
	});
	const ago = $derived(savedAt ? formatAgo(savedAt, savedAt > now ? savedAt : now) : '');
</script>

<!-- 화면 낭독기에는 조용히 알린다 (입력을 끊지 않게) -->
<span class="flex min-h-11 items-center gap-1.5 text-xs" aria-live="polite">
	{#if status === 'saving'}
		<span class="text-muted-foreground">저장 중…</span>
	{:else if status === 'saved' && ago}
		<span class="text-muted-foreground">저장됨 · {ago}</span>
	{:else if status === 'failed' || status === 'limit'}
		<CircleAlert class="size-3.5 shrink-0" aria-hidden="true" />
		<span class="font-semibold">저장 안 됨</span>
		<button
			type="button"
			class="min-h-11 px-1 text-accent-text underline underline-offset-2"
			onclick={onretry}>다시 시도</button
		>
	{/if}
</span>
