<script lang="ts">
	// S-02 오늘 (DESIGN 4.1절): 날짜 · 질문 → 체크인 원 → 오늘 기록(시간순, 일기 형식).
	// "오늘"은 서버가 사용자 시간대로 정한 날짜를 그대로 쓴다 — 화면과 서버의 기준이 어긋나지 않게 (DATA_MODEL 3.4절).
	// 이어 쓰기(초안) · 사고 기록 쓰기 버튼은 M5에서, 서버 상태 확인(M1 줄기)은 /api/health로 충분해서 이 화면에서 뺐다.
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import CircleAlert from '@lucide/svelte/icons/circle-alert';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import CheckinOrb from '#lib/components/CheckinOrb.svelte';
	import IntensityDots from '#lib/components/IntensityDots.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { formatTime } from '#lib/tz.js';

	type CheckinDay = components['schemas']['CheckinDay'];

	const tz = $derived(auth.user?.timezone ?? 'Asia/Seoul');
	let day = $state<CheckinDay | null>(null);
	let loadError = $state('');
	let loading = $state(true);

	// 서버가 준 날짜(YYYY-MM-DD)를 "2026년 11월 1일 일요일"로. 받기 전에는 지금 시각을 사용자 시간대로.
	const dateText = $derived.by(() => {
		const fmt = (d: Date, timeZone: string) =>
			new Intl.DateTimeFormat('ko-KR', { dateStyle: 'full', timeZone }).format(d);
		if (!day) return fmt(new Date(), tz);
		const [y, m, d] = day.date.split('-').map(Number);
		// 그 날짜의 정오(UTC)를 UTC로 읽으면 어느 시간대에서도 날짜가 밀리지 않는다
		return fmt(new Date(Date.UTC(y, m - 1, d, 12)), 'UTC');
	});

	async function load() {
		loading = true;
		loadError = '';
		// 이미 보이는 목록은 다시 불러오는 동안에도 지우지 않는다 (DESIGN 7장)
		try {
			const { data, error } = await api.GET('/api/checkins');
			if (data) day = data;
			else loadError = errorMessage(error);
		} catch {
			loadError = NETWORK_ERROR;
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<!-- 머리: 날짜 · 질문은 명조 (DESIGN 4.1절 · 6장 "서체") -->
<header class="mb-8 text-center">
	<p class="font-serif text-muted-foreground">{dateText}</p>
	<h1 class="mt-2 font-serif text-3xl leading-snug font-bold">오늘은 어떤 하루였나요?</h1>
</header>

<div class="mb-10">
	<CheckinOrb />
</div>

<!-- 오늘: 왼쪽 여백에 시각, 강도는 숫자 + 점 5개 (DESIGN 4.1절). 줄을 누르면 고치기(C4) -->
<section class="border-t pt-5" aria-labelledby="today-title">
	<h2 id="today-title" class="mb-1 font-serif text-lg font-bold">오늘</h2>

	{#if day && day.items.length > 0}
		<ul>
			{#each day.items as c (c.id)}
				<li class="border-b border-dashed">
					<a
						href={resolve('/(app)/checkin/[id]/edit', { id: c.id })}
						class="grid grid-cols-[3.5rem_1fr] gap-2.5 py-3.5 hover:bg-muted/40"
					>
						<time
							datetime={c.occurred_at}
							class="pt-px font-serif text-muted-foreground tabular-nums"
							>{formatTime(c.occurred_at, tz)}</time
						>
						<span class="space-y-0.5 leading-relaxed">
							{#each c.emotions as e (e.emotion_id)}
								<span class="block">
									{e.name} <span class="tabular-nums">{e.intensity}</span>
									<IntensityDots value={e.intensity} />
								</span>
							{/each}
							{#if c.note}<span class="block text-sm text-muted-foreground">{c.note}</span>{/if}
						</span>
					</a>
				</li>
			{/each}
		</ul>
	{:else if day}
		<!-- 첫 사용 · 오늘 0건 (DESIGN 7장) — 버튼은 위의 체크인 원 -->
		<p class="py-4 text-muted-foreground">첫 체크인을 해 볼까요?</p>
	{:else if loadError}
		<div class="space-y-3 py-3">
			<!-- 오류는 빨강이 아니라 아이콘 + 굵은 글자 (DESIGN 6장 "오류 표시") -->
			<p class="flex items-center gap-2 font-semibold">
				<CircleAlert class="size-4" aria-hidden="true" />
				{loadError}
			</p>
			<Button variant="outline" class="min-h-11 border-input" onclick={load} disabled={loading}
				>다시 시도</Button
			>
		</div>
	{:else}
		<p class="py-4 text-muted-foreground">불러오는 중…</p>
	{/if}
</section>
