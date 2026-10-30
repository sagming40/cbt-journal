<script lang="ts">
	// 오류 문구: 빨강을 쓰지 않고 아이콘 + 본문 색 굵은 글자로 알린다 (DESIGN 6장 "오류 표시").
	// 빨강이 "나쁜 감정"이나 경고로 읽히는 걸 피한다.
	import CircleAlert from '@lucide/svelte/icons/circle-alert';

	let { message, id }: { message?: string; id?: string } = $props();
</script>

{#if message}
	<p {id} role="alert" class="flex items-start gap-1.5 text-sm font-semibold">
		<CircleAlert class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
		<span>{message}</span>
	</p>
{/if}
