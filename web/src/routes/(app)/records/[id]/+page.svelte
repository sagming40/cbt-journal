<script lang="ts">
	// S-06 사고 기록 상세 (DESIGN 4.5절). 칸을 작성 순서대로, 감정은 전 → 후 숫자만(색 없음).
	// 없거나 남의 것이면 "찾을 수 없는 기록이에요" (DESIGN 7장 404).
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import X from '@lucide/svelte/icons/x';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import FormError from '#lib/components/FormError.svelte';
	import IntensityChange from '#lib/components/IntensityChange.svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { formatDateTime } from '#lib/tz.js';

	type ThoughtRecord = components['schemas']['Record'];

	const id = $derived(page.params.id ?? '');
	const tz = $derived(auth.user?.timezone ?? 'Asia/Seoul');
	let rec = $state<ThoughtRecord | null>(null);
	let notFound = $state(false);
	let loadError = $state('');
	let deleting = $state(false);
	let deleteMessage = $state('');

	async function load() {
		loadError = '';
		try {
			const { data, error, response } = await api.GET('/api/records/{id}', {
				params: { path: { id } }
			});
			if (data) rec = data;
			else if (response.status === 404) notFound = true;
			else loadError = errorMessage(error);
		} catch {
			loadError = NETWORK_ERROR;
		}
	}

	onMount(load);

	async function remove() {
		deleting = true;
		deleteMessage = '';
		try {
			const { error, response } = await api.DELETE('/api/records/{id}', {
				params: { path: { id } }
			});
			// 이미 지워져 있어도(404) 사용자가 원한 결과는 같다
			if (response.status === 204 || response.status === 404) {
				toast.success('지웠어요');
				await goto(resolve('/(app)'), { replace: true });
				return;
			}
			deleteMessage = errorMessage(error);
		} catch {
			deleteMessage = NETWORK_ERROR;
		} finally {
			deleting = false;
		}
	}

	const editHref = (step: number) => `${resolve('/(app)/records/[id]/edit', { id })}?step=${step}`;
</script>

<svelte:head><title>사고 기록 · CBT 저널</title></svelte:head>

{#snippet section(title: string, step: number, text: string)}
	<section class="border-b border-dashed py-5 first:pt-0">
		<div class="flex items-baseline justify-between gap-3">
			<h2 class="font-serif text-base font-bold">{title}</h2>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- editHref가 resolve로 만든 주소다 -->
			<a
				href={editHref(step)}
				class="text-xs text-muted-foreground underline-offset-4 hover:underline"
				>고치기<span class="sr-only"> ({title})</span></a
			>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
		{#if text}
			<p class="mt-2 leading-relaxed whitespace-pre-wrap">{text}</p>
		{:else}
			<p class="mt-2 text-sm text-muted-foreground">비워 뒀어요</p>
		{/if}
	</section>
{/snippet}

<header class="mb-6 flex items-center gap-2 md:mb-9 md:border-b md:pb-4">
	<a
		href={resolve('/(app)')}
		class="-ml-2 flex size-11 shrink-0 items-center justify-center rounded-md hover:bg-muted"
		aria-label="닫기"
	>
		<X class="size-5" aria-hidden="true" />
	</a>
	<div class="min-w-0 flex-1">
		<h1 class="font-serif text-lg font-bold">사고 기록</h1>
		{#if rec}
			<p class="text-sm text-muted-foreground tabular-nums">
				{formatDateTime(rec.occurred_at, tz)}
			</p>
		{/if}
	</div>
	{#if rec}
		<Button href={editHref(1)} variant="outline" class="min-h-11 border-input">수정</Button>
		<AlertDialog.Root>
			<AlertDialog.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="ghost" class="min-h-11">지우기</Button>
				{/snippet}
			</AlertDialog.Trigger>
			<AlertDialog.Content class="rounded-md">
				<AlertDialog.Header>
					<AlertDialog.Title class="font-serif">이 기록을 지울까요?</AlertDialog.Title>
					<AlertDialog.Description>지운 기록은 되돌릴 수 없어요.</AlertDialog.Description>
				</AlertDialog.Header>
				<AlertDialog.Footer>
					<AlertDialog.Cancel class="min-h-11 border-input">그대로 두기</AlertDialog.Cancel>
					<AlertDialog.Action class="min-h-11" onclick={remove} disabled={deleting}>
						{deleting ? '지우는 중…' : '지우기'}
					</AlertDialog.Action>
				</AlertDialog.Footer>
			</AlertDialog.Content>
		</AlertDialog.Root>
	{/if}
</header>

{#if deleteMessage}<div class="mb-4"><FormError message={deleteMessage} /></div>{/if}

{#if notFound}
	<div class="space-y-4 py-6">
		<FormError message="찾을 수 없는 기록이에요" />
		<a href={resolve('/(app)')} class="text-accent-text underline-offset-4 hover:underline"
			>오늘로 돌아가기</a
		>
	</div>
{:else if rec}
	<!-- 데스크톱(64rem 이상) 2단: 왼쪽 글 칸들 | 오른쪽 종이에 감정 · 왜곡 (DESIGN 4.5절). 모바일은 작성 순서대로 -->
	<div class="lg:grid lg:grid-cols-[minmax(0,40rem)_20rem] lg:items-start lg:gap-10">
		<div class="contents lg:block">
			{@render section('상황', 1, rec.situation)}
			<div class="lg:hidden">{@render emotionsBlock(rec)}</div>
			{@render section('자동 사고', 3, rec.automatic_thought)}
			<div class="lg:hidden">{@render distortionsBlock(rec)}</div>
			{@render section('근거', 5, rec.evidence_for)}
			{@render section('반대 근거', 6, rec.evidence_against)}
			{@render section('대안 사고', 7, rec.alternative_thought)}
		</div>
		<aside
			class="hidden rounded-lg border bg-card px-6 pt-5 pb-2 lg:sticky lg:top-8 lg:block [&>section:first-child]:pt-0 [&>section:last-child]:border-b-0"
		>
			{@render emotionsBlock(rec)}
			{@render distortionsBlock(rec)}
		</aside>
	</div>
{:else if loadError}
	<div class="space-y-4">
		<FormError message={loadError} />
		<Button variant="outline" class="min-h-11 border-input" onclick={load}>다시 시도</Button>
	</div>
{:else}
	<p class="text-muted-foreground">불러오는 중…</p>
{/if}

{#snippet emotionsBlock(r: ThoughtRecord)}
	<section class="border-b border-dashed py-5">
		<div class="flex items-baseline justify-between gap-3">
			<h2 class="font-serif text-base font-bold">감정 전 → 후</h2>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- editHref가 resolve로 만든 주소다 -->
			<a href={editHref(8)} class="text-xs text-muted-foreground underline-offset-4 hover:underline"
				>고치기<span class="sr-only"> (감정)</span></a
			>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
		<ul class="mt-2 space-y-1">
			{#each r.emotions as e (e.emotion_id)}
				<li class="flex items-baseline justify-between gap-3">
					<span class="font-serif font-bold">{e.name}</span>
					<IntensityChange before={e.intensity_before} after={e.intensity_after} />
				</li>
			{/each}
		</ul>
	</section>
{/snippet}

{#snippet distortionsBlock(r: ThoughtRecord)}
	<section class="border-b border-dashed py-5">
		<div class="flex items-baseline justify-between gap-3">
			<h2 class="font-serif text-base font-bold">인지 왜곡</h2>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- editHref가 resolve로 만든 주소다 -->
			<a href={editHref(4)} class="text-xs text-muted-foreground underline-offset-4 hover:underline"
				>고치기<span class="sr-only"> (인지 왜곡)</span></a
			>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
		{#if r.distortions.length > 0}
			<ul class="mt-2 flex flex-wrap gap-1.5">
				{#each r.distortions as d (d.id)}
					<li class="rounded-md border px-2.5 py-1 text-sm">{d.name}</li>
				{/each}
			</ul>
		{:else}
			<p class="mt-2 text-sm text-muted-foreground">고르지 않았어요</p>
		{/if}
	</section>
{/snippet}
