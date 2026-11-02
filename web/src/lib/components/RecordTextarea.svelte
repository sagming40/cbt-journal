<script lang="ts">
	// 사고 기록의 긴 글 칸 — rule 줄무늬가 깔린 공책 칸 (DESIGN 4.3절 "칸 표시", 6장 "모양 D2").
	// 줄 간격 32px과 줄무늬 주기를 맞춘다. 글이 길어지면 칸도 같이 늘어난다(지원하지 않는 브라우저는 안에서 스크롤).
	import FormError from '#lib/components/FormError.svelte';

	let {
		id,
		label,
		value = $bindable(''),
		maxlength,
		error,
		placeholder
	}: {
		id: string;
		label: string;
		value?: string;
		maxlength: number;
		error?: string;
		placeholder?: string;
	} = $props();

	// 다 찼을 때 갑자기 입력이 막히지 않게, 끝이 가까우면 남은 양을 보여 준다
	const nearLimit = $derived(value.length >= maxlength * 0.9);
</script>

<div>
	<label for={id} class="text-sm text-muted-foreground">{label}</label>
	<textarea
		{id}
		bind:value
		{maxlength}
		{placeholder}
		rows="6"
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={error ? `${id}-error` : undefined}
		class="lined mt-1 block min-h-48 w-full resize-none border-0 border-b border-input bg-transparent px-0.5 text-[17px] leading-8 outline-none placeholder:text-muted-foreground focus-visible:border-b-2 focus-visible:border-primary"
	></textarea>
	{#if nearLimit}
		<p class="mt-1 text-right text-xs text-muted-foreground tabular-nums">
			{value.length.toLocaleString('ko-KR')} / {maxlength.toLocaleString('ko-KR')}
		</p>
	{/if}
	<div class="mt-1"><FormError id={`${id}-error`} message={error} /></div>
</div>

<style>
	.lined {
		field-sizing: content;
		background-image: repeating-linear-gradient(
			to bottom,
			transparent 0 31px,
			var(--rule) 31px 32px
		);
		background-attachment: local;
	}
</style>
