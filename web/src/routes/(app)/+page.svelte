<script lang="ts">
	// S-02 오늘 (DESIGN 4.1절): 날짜 · 질문 → 체크인 원 → 오늘 기록(시간순, 일기 형식).
	// "오늘"은 서버가 사용자 시간대로 정한 날짜를 그대로 쓴다 — 화면과 서버의 기준이 어긋나지 않게 (DATA_MODEL 3.4절).
	// 오늘 목록은 체크인 · 사고 기록을 시간순으로 섞는다. 서버 상태 확인(M1 줄기)은 /api/health로 충분해서 이 화면에서 뺐다.
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import CircleAlert from '@lucide/svelte/icons/circle-alert';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import CheckinOrb from '#lib/components/CheckinOrb.svelte';
	import Logo from '#lib/components/Logo.svelte';
	import IntensityDots from '#lib/components/IntensityDots.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import IntensityChange from '#lib/components/IntensityChange.svelte';
	import ResumeDrafts from '#lib/components/ResumeDrafts.svelte';
	import { formatTime } from '#lib/tz.js';

	type CheckinDay = components['schemas']['CheckinDay'];
	type RecordDay = components['schemas']['RecordDay'];
	type Checkin = components['schemas']['Checkin'];
	type ThoughtRecord = components['schemas']['Record'];
	type Row = { kind: 'checkin'; c: Checkin } | { kind: 'record'; r: ThoughtRecord };

	const tz = $derived(auth.user?.timezone ?? 'Asia/Seoul');
	let day = $state<CheckinDay | null>(null);
	let records = $state<RecordDay | null>(null);
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
			const [ck, rc] = await Promise.all([api.GET('/api/checkins'), api.GET('/api/records')]);
			if (ck.data && rc.data) {
				day = ck.data;
				records = rc.data;
			} else {
				loadError = errorMessage(ck.error ?? rc.error);
			}
		} catch {
			loadError = NETWORK_ERROR;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	// 체크인 · 사고 기록을 일어난 시각순으로 (DESIGN 4.1절 "오늘 목록")
	const rows = $derived.by((): Row[] => {
		const list: Row[] = [
			...(day?.items ?? []).map((c): Row => ({ kind: 'checkin', c })),
			...(records?.items ?? []).map((r): Row => ({ kind: 'record', r }))
		];
		const at = (x: Row) => new Date(x.kind === 'checkin' ? x.c.occurred_at : x.r.occurred_at);
		return list.sort((a, b) => at(a).getTime() - at(b).getTime());
	});

	// 감정이 여러 개면 변화가 가장 큰 감정 하나 + "외 2" (DESIGN 4.4절 "사고 기록 카드"와 같게)
	function biggest(r: ThoughtRecord) {
		return r.emotions.reduce((a, b) =>
			Math.abs(b.intensity_after - b.intensity_before) >
			Math.abs(a.intensity_after - a.intensity_before)
				? b
				: a
		);
	}
</script>

<!-- 머리: 날짜 · 질문은 명조 (DESIGN 4.1절 · 6장 "서체"). 데스크톱은 왼쪽 정렬 + 아래 선 (3.2절) -->
<header class="mb-8 text-center md:mb-10 md:border-b md:pb-6 md:text-left">
	<p class="font-serif text-muted-foreground">{dateText}</p>
	<h1 class="mt-2 font-serif text-3xl leading-snug font-bold md:text-[34px]">
		오늘은 어떤 하루였나요?
	</h1>
</header>

<!-- 데스크톱 2단(64rem 이상): 왼쪽 종이 = 지금 할 일, 오른쪽 = 한 일 (DESIGN 4.1절 "데스크톱"). 모바일은 위아래로 -->
<div class="lg:grid lg:grid-cols-[22rem_1fr] lg:items-start lg:gap-10 xl:grid-cols-[24rem_1fr]">
	<div class="mb-10 md:rounded-lg md:border md:bg-card md:px-7 md:pt-10 md:pb-9 lg:mb-0">
		<CheckinOrb />
		<div class="mt-6 flex justify-center">
			<Button
				href={resolve('/(app)/records/new')}
				variant="outline"
				class="min-h-11 border-input px-6">사고 기록 쓰기</Button
			>
		</div>
		<ResumeDrafts />
	</div>

	<!-- 오늘: 왼쪽 여백에 시각, 강도는 숫자 + 점 5개 (DESIGN 4.1절). 줄을 누르면 고치기(C4) -->
	<section class="border-t pt-5 lg:border-t-0 lg:pt-0" aria-labelledby="today-title">
		<h2 id="today-title" class="mb-1 font-serif text-lg font-bold md:mb-2 md:border-b md:pb-2">
			오늘
		</h2>

		{#if day && rows.length > 0}
			<ul>
				{#each rows as row (row.kind === 'checkin' ? row.c.id : row.r.id)}
					<li class="border-b border-dashed">
						{#if row.kind === 'record'}
							{@const top = biggest(row.r)}
							<a
								href={resolve('/(app)/records/[id]', { id: row.r.id })}
								class="grid grid-cols-[3.5rem_1fr] gap-2.5 py-3.5 hover:bg-muted/40 md:grid-cols-[4rem_1fr] md:gap-3.5 md:px-1 md:py-4"
							>
								<time
									datetime={row.r.occurred_at}
									class="pt-px font-serif text-muted-foreground tabular-nums"
									>{formatTime(row.r.occurred_at, tz)}</time
								>
								<span class="space-y-0.5 leading-relaxed">
									<span class="block text-sm text-muted-foreground">사고 기록</span>
									<span class="block">
										<span class="font-serif font-bold">{top.name}</span>
										<IntensityChange before={top.intensity_before} after={top.intensity_after} />
										{#if row.r.emotions.length > 1}<span class="text-sm text-muted-foreground"
												>외 {row.r.emotions.length - 1}</span
											>{/if}
									</span>
								</span>
							</a>
						{:else}
							{@const c = row.c}
							<a
								href={resolve('/(app)/checkin/[id]/edit', { id: c.id })}
								class="grid grid-cols-[3.5rem_1fr] gap-2.5 py-3.5 hover:bg-muted/40 md:grid-cols-[4rem_1fr] md:gap-3.5 md:px-1 md:py-4"
							>
								<time
									datetime={c.occurred_at}
									class="pt-px font-serif text-muted-foreground tabular-nums"
									>{formatTime(c.occurred_at, tz)}</time
								>
								<span class="space-y-0.5 leading-relaxed">
									{#each c.emotions as e (e.emotion_id)}
										<span class="block">
											<span class="font-serif font-bold">{e.name}</span>
											<span class="tabular-nums">{e.intensity}</span>
											<IntensityDots value={e.intensity} />
										</span>
									{/each}
									{#if c.note}<span class="block text-sm text-muted-foreground">{c.note}</span>{/if}
								</span>
							</a>
						{/if}
					</li>
				{/each}
			</ul>
		{:else if day}
			<!-- 첫 사용 · 오늘 0건 (DESIGN 7장) — 버튼은 체크인 원. 데스크톱은 빈자리에 로고의 두 원을 흐리게 (4.1절) -->
			<div
				class="py-4 md:flex md:min-h-72 md:flex-col md:items-center md:justify-center md:gap-3 md:text-center"
			>
				<span class="hidden opacity-30 md:block" aria-hidden="true"><Logo size={64} /></span>
				<p class="text-muted-foreground md:font-serif md:text-lg md:text-foreground">
					첫 체크인을 해 볼까요?
				</p>
				<p class="hidden text-sm text-muted-foreground md:block">
					지금 마음이 어떤지, 원을 눌러 한 번 적어 봐요
				</p>
			</div>
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
</div>
