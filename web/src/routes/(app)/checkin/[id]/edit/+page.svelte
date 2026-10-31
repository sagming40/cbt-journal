<script lang="ts">
	// S-03 체크인 수정 (DESIGN 4.2절 "수정 모드", REQUIREMENTS C4). 같은 입력 화면에 기존 값을 채워 연다.
	// 헤더에 삭제(D-01 확인 창). 없거나 남의 것이면 "찾을 수 없는 기록이에요" (DESIGN 7장 404).
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import X from '@lucide/svelte/icons/x';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import CheckinForm from '#lib/components/CheckinForm.svelte';
	import FormError from '#lib/components/FormError.svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';

	type Emotion = components['schemas']['Emotion'];
	type Checkin = components['schemas']['Checkin'];
	type CheckinInput = components['schemas']['CheckinInput'];

	const id = $derived(page.params.id ?? '');
	let emotions = $state<Emotion[] | null>(null);
	let checkin = $state<Checkin | null>(null);
	let notFound = $state(false);
	let loadError = $state('');
	let submitting = $state(false);
	let fields = $state<Record<string, string>>({});
	let message = $state('');
	let deleting = $state(false);
	let deleteMessage = $state('');

	async function load() {
		loadError = '';
		try {
			const [em, ck] = await Promise.all([
				api.GET('/api/emotions'),
				api.GET('/api/checkins/{id}', { params: { path: { id } } })
			]);
			if (ck.response.status === 404) {
				notFound = true;
				return;
			}
			if (em.data && ck.data) {
				emotions = em.data.items;
				checkin = ck.data;
			} else {
				loadError = errorMessage(em.error ?? ck.error);
			}
		} catch {
			loadError = NETWORK_ERROR;
		}
	}

	onMount(load);

	async function save(body: CheckinInput) {
		submitting = true;
		fields = {};
		message = '';
		try {
			const { data, error, response } = await api.PUT('/api/checkins/{id}', {
				params: { path: { id } },
				body
			});
			if (data) {
				toast.success('고쳤어요');
				await goto(resolve('/(app)'));
				return;
			}
			if (response.status === 404) {
				// 다른 기기에서 지운 뒤에 고친 경우 (DATA_MODEL 3.4절)
				notFound = true;
				return;
			}
			fields = fieldErrors(error);
			message = Object.keys(fields).length > 0 ? '' : errorMessage(error);
		} catch {
			message = NETWORK_ERROR;
		} finally {
			submitting = false;
		}
	}

	async function remove() {
		deleting = true;
		deleteMessage = '';
		try {
			const { error, response } = await api.DELETE('/api/checkins/{id}', {
				params: { path: { id } }
			});
			// 이미 지워져 있어도(404) 사용자가 원한 결과는 같다 — 오늘로 돌아간다
			if (response.status === 204 || response.status === 404) {
				toast.success('지웠어요');
				await goto(resolve('/(app)'));
				return;
			}
			deleteMessage = errorMessage(error);
		} catch {
			deleteMessage = NETWORK_ERROR;
		} finally {
			deleting = false;
		}
	}
</script>

<svelte:head><title>체크인 고치기 · CBT 저널</title></svelte:head>

<header class="mb-6 flex items-center gap-2">
	<a
		href={resolve('/(app)')}
		class="-ml-2 flex size-11 items-center justify-center rounded-md hover:bg-muted"
		aria-label="닫기"
	>
		<X class="size-5" aria-hidden="true" />
	</a>
	<h1 class="flex-1 font-serif text-lg font-bold">체크인 고치기</h1>
	{#if checkin}
		<!-- D-01 삭제 확인 (DESIGN 2장). 되돌릴 수 없어서 한 번 더 묻는다 -->
		<AlertDialog.Root>
			<AlertDialog.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="ghost" class="min-h-11">지우기</Button>
				{/snippet}
			</AlertDialog.Trigger>
			<!-- 모서리 6px · 그림자 없음 (DESIGN 6장 "모양 D2") — 생성 부품은 고치지 않고 class로 덮는다 -->
			<AlertDialog.Content class="rounded-md">
				<AlertDialog.Header>
					<AlertDialog.Title class="font-serif">이 체크인을 지울까요?</AlertDialog.Title>
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

<!-- [지우기]를 누르면 확인 창이 바로 닫히므로, 지우기가 실패한 문구는 창 밖(화면)에 보인다 -->
{#if deleteMessage}<div class="mb-4"><FormError message={deleteMessage} /></div>{/if}

{#if notFound}
	<div class="space-y-4 py-6">
		<FormError message="찾을 수 없는 기록이에요" />
		<a href={resolve('/(app)')} class="text-accent-text underline-offset-4 hover:underline"
			>오늘로 돌아가기</a
		>
	</div>
{:else if emotions && checkin}
	<CheckinForm
		{emotions}
		initial={{
			occurred_at: checkin.occurred_at,
			note: checkin.note,
			emotions: checkin.emotions.map((e) => ({ emotion_id: e.emotion_id, intensity: e.intensity }))
		}}
		{submitting}
		{fields}
		{message}
		onsubmit={save}
	/>
{:else if loadError}
	<div class="space-y-4">
		<FormError message={loadError} />
		<Button variant="outline" class="min-h-11 border-input" onclick={load}>다시 시도</Button>
	</div>
{:else}
	<p class="text-muted-foreground">불러오는 중…</p>
{/if}
