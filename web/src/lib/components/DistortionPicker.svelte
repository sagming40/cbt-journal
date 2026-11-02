<script lang="ts">
	// 인지 왜곡 고르기 (DESIGN 4.3절 4단계, REQUIREMENTS T4). 여러 개, 0개도 된다.
	// 줄을 누르면 고르고, 설명이 있으면 오른쪽 [설명]으로 펼친다. 설명 문구는 사용자가 직접 쓴다 — 비어 있으면 이름만.
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import type { components } from '#lib/api/schema.js';

	type Distortion = components['schemas']['Distortion'];

	let { distortions, value = $bindable([]) }: { distortions: Distortion[]; value?: number[] } =
		$props();

	let open = $state<number | null>(null);

	function toggle(id: number) {
		value = value.includes(id) ? value.filter((x) => x !== id) : [...value, id];
	}
</script>

<fieldset>
	<legend class="sr-only">인지 왜곡</legend>
	<ul class="border-t border-dashed">
		{#each distortions as d (d.id)}
			{@const picked = value.includes(d.id)}
			{@const hasText = Boolean(d.description || d.example)}
			<li class="border-b border-dashed">
				<div class="flex items-stretch gap-1">
					<button
						type="button"
						aria-pressed={picked}
						onclick={() => toggle(d.id)}
						class={[
							'my-1 min-h-11 flex-1 rounded-md border px-3 py-2 text-left text-[15px] transition-colors',
							picked
								? 'border-primary bg-accent-tint font-bold text-accent-text shadow-[inset_0_-3px_0_var(--primary)] dark:border-accent-text dark:shadow-[inset_0_-3px_0_var(--accent-text)]'
								: 'border-transparent hover:bg-muted'
						]}>{d.name}</button
					>
					{#if hasText}
						<button
							type="button"
							class="flex min-h-11 w-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
							aria-expanded={open === d.id}
							aria-label={`${d.name} 설명`}
							onclick={() => (open = open === d.id ? null : d.id)}
						>
							<ChevronDown
								class={['size-4 transition-transform', open === d.id && 'rotate-180']}
								aria-hidden="true"
							/>
						</button>
					{/if}
				</div>
				{#if open === d.id}
					<div class="space-y-1 px-3 pb-3 text-sm leading-relaxed">
						{#if d.description}<p>{d.description}</p>{/if}
						{#if d.example}<p class="text-muted-foreground">예: {d.example}</p>{/if}
					</div>
				{/if}
			</li>
		{/each}
	</ul>
</fieldset>
