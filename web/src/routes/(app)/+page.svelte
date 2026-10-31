<script lang="ts">
	// S-02 오늘 자리. 체크인 원(CheckinOrb) · 오늘 기록은 M4 체크인에서 들어온다(누르면 갈 /checkin이 그때 생긴다).
	// 지금은 날짜 · 질문(D2 머리)과 서버 상태(M1 줄기 확인)만.
	import { onMount } from 'svelte';
	import CircleAlert from '@lucide/svelte/icons/circle-alert';
	import CircleCheck from '@lucide/svelte/icons/circle-check';
	import { api } from '#lib/api/client.js';
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import { Button } from '#lib/components/ui/button/index.js';

	type Health = components['schemas']['Health'];

	let health = $state<Health | null>(null);
	let unreachable = $state(false); // 서버에 아예 닿지 못함 (Go가 꺼짐 등)
	let loading = $state(true);

	// "오늘"은 사용자 시간대로 (CLAUDE.md "시각은 UTC로 저장, 하루는 users.timezone"). 예: 2026년 11월 1일 일요일
	const today = $derived(
		new Intl.DateTimeFormat('ko-KR', { dateStyle: 'full', timeZone: auth.user?.timezone }).format(
			new Date()
		)
	);

	async function load() {
		loading = true;
		// 이전 결과는 지우지 않는다: 다시 확인하는 동안에도 보이게 둔다 (DESIGN 7장).
		try {
			const { data, error, response } = await api.GET('/api/health');
			// 계약서에 적힌 200 · 503만 Health 모양이다. 그 밖(프록시 502 등)은 서버에 닿지 못한 것으로 본다.
			const body = data ?? error;
			if ((response.status === 200 || response.status === 503) && body) {
				health = body;
				unreachable = false;
			} else {
				health = null;
				unreachable = true;
			}
		} catch {
			health = null;
			unreachable = true;
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<!-- 머리: 날짜 · 질문은 명조 (DESIGN 4.1절 · 6장 "서체") -->
<header class="mb-10 text-center">
	<p class="font-serif text-muted-foreground">{today}</p>
	<h1 class="mt-2 font-serif text-3xl leading-snug font-bold">오늘은 어떤 하루였나요?</h1>
	<p class="mt-3 text-sm text-muted-foreground">
		{auth.user?.nickname}님, 체크인과 사고 기록은 곧 여기에 생겨요
	</p>
</header>

<!-- 구역은 상자 대신 선으로 나눈다 (DESIGN 6장 "모양 D2") -->
<section class="border-t pt-5" aria-labelledby="health-title">
	<div class="mb-1 flex items-baseline justify-between">
		<h2 id="health-title" class="font-serif text-lg font-bold">서버 상태</h2>
		<span class="text-xs text-muted-foreground">브라우저 → Go → Postgres · Valkey</span>
	</div>

	{#if loading && !health && !unreachable}
		<p class="py-3 text-muted-foreground">확인하는 중…</p>
	{:else if unreachable}
		<!-- 오류는 빨강이 아니라 아이콘 + 굵은 글자 (DESIGN 6장 "오류 표시") -->
		<p class="flex items-center gap-2 py-3 font-semibold">
			<CircleAlert class="size-4" aria-hidden="true" /> 연결이 안 돼요
		</p>
	{:else if health}
		<ul>
			{#each [{ name: 'Postgres', status: health.postgres }, { name: 'Valkey', status: health.valkey }] as item (item.name)}
				<li class="flex items-center justify-between border-b border-dashed py-3">
					<span>{item.name}</span>
					{#if item.status === 'ok'}
						<span class="flex items-center gap-1 text-accent-text">
							<CircleCheck class="size-4" aria-hidden="true" /> ok
						</span>
					{:else}
						<span class="flex items-center gap-1 font-semibold">
							<CircleAlert class="size-4" aria-hidden="true" /> error
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}

	<Button variant="outline" class="mt-4 min-h-11 border-input" onclick={load} disabled={loading}
		>다시 시도</Button
	>
</section>
