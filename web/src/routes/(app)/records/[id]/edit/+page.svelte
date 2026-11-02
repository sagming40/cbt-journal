<script lang="ts">
	// S-04 수정 모드 (DESIGN 4.3절, REQUIREMENTS T5): 초안 · 자동 저장 없이 [저장] 한 번.
	// 나중에 저장한 쪽이 이긴다(DATA_MODEL 3.5절). 지워진 걸 고치면 "찾을 수 없는 기록이에요".
	import { onMount } from 'svelte';
	import { afterNavigate, beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import FormError from '#lib/components/FormError.svelte';
	import RecordStepFields from '#lib/components/RecordStepFields.svelte';
	import StepShell from '#lib/components/StepShell.svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import {
		LAST_STEP,
		canFinish,
		clampStep,
		emptyForm,
		firstErrorStep,
		pickForm,
		stepComplete,
		toRecordInput,
		type RecordForm
	} from '#lib/record-form.js';

	type Emotion = components['schemas']['Emotion'];
	type Distortion = components['schemas']['Distortion'];

	const id = $derived(page.params.id ?? '');
	let emotions = $state<Emotion[] | null>(null);
	let distortions = $state<Distortion[] | null>(null);
	let form = $state<RecordForm>(emptyForm(new Date()));
	let ready = $state(false);
	let notFound = $state(false);
	let loadError = $state('');
	let fields = $state<Record<string, string>>({});
	let message = $state('');
	let saving = $state(false);
	let original = '';
	let leaveTo = $state<URL | null>(null);
	let leaving = false;

	// 이동할 때만 주소를 다시 읽는다 — 칸을 비우는 것만으로 단계가 바뀌면 안 된다
	let step = $state(1);
	afterNavigate(() => {
		if (ready) step = clampStep(Number(page.url.searchParams.get('step') ?? 1), form);
	});
	const detailHref = $derived(resolve('/(app)/records/[id]', { id }));
	const dirty = $derived(ready && JSON.stringify(form) !== original);

	async function load() {
		loadError = '';
		try {
			const [em, ds, rec] = await Promise.all([
				api.GET('/api/emotions'),
				api.GET('/api/distortions'),
				api.GET('/api/records/{id}', { params: { path: { id } } })
			]);
			if (rec.response.status === 404) {
				notFound = true;
				return;
			}
			if (!em.data || !ds.data || !rec.data) {
				loadError = errorMessage(em.error ?? ds.error ?? rec.error);
				return;
			}
			emotions = em.data.items;
			distortions = ds.data.items;
			form = pickForm({
				...rec.data,
				distortion_ids: rec.data.distortions.map((d) => d.id)
			});
			original = JSON.stringify(form);
			step = clampStep(Number(page.url.searchParams.get('step') ?? 1), form);
			ready = true;
		} catch {
			loadError = NETWORK_ERROR;
		}
	}

	onMount(load);

	function go(s: number) {
		fields = {};
		message = '';
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve로 만든 주소에 단계만 붙였다
		goto(`${resolve('/(app)/records/[id]/edit', { id })}?step=${s}`);
	}

	async function save() {
		if (saving) return;
		saving = true;
		fields = {};
		message = '';
		try {
			const { data, error, response } = await api.PUT('/api/records/{id}', {
				params: { path: { id } },
				body: toRecordInput($state.snapshot(form) as RecordForm)
			});
			if (data) {
				leaving = true;
				toast.success('고쳤어요');
				await goto(detailHref, { replace: true });
				return;
			}
			if (response.status === 404) {
				// 다른 기기에서 지운 뒤에 고친 경우 (DATA_MODEL 3.5절)
				notFound = true;
				return;
			}
			fields = fieldErrors(error);
			const at = firstErrorStep(fields);
			if (at !== null) go(at);
			else message = errorMessage(error);
		} catch {
			message = NETWORK_ERROR;
		} finally {
			saving = false;
		}
	}

	// 고친 내용을 잃지 않게 (DESIGN 4.3절 "나가기"). 자동 저장이 없는 화면이라 고친 게 있으면 묻는다.
	beforeNavigate((nav) => {
		if (leaving || !dirty || !nav.to || nav.to.url.pathname === page.url.pathname) return;
		nav.cancel();
		if (nav.type !== 'leave') leaveTo = nav.to.url;
	});

	function leaveAnyway() {
		const to = leaveTo;
		leaveTo = null;
		leaving = true;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- 앱 안에서 누른 링크의 주소 그대로다
		if (to) goto(to);
	}

	const canSave = $derived(!saving && canFinish(form));
</script>

<svelte:head><title>사고 기록 고치기 · CBT 저널</title></svelte:head>

{#if notFound}
	<div class="space-y-4 py-6">
		<FormError message="찾을 수 없는 기록이에요" />
		<a href={resolve('/(app)')} class="text-accent-text underline-offset-4 hover:underline"
			>오늘로 돌아가기</a
		>
	</div>
{:else if ready && emotions && distortions}
	<StepShell
		heading="사고 기록 고치기"
		closeHref={detailHref}
		{step}
		{form}
		{emotions}
		{distortions}
		{message}
		nextLabel={step === LAST_STEP ? (saving ? '저장하는 중…' : '저장') : '다음'}
		nextDisabled={step === LAST_STEP ? !canSave : !stepComplete(step, form)}
		onprev={() => go(step - 1)}
		onnext={() => (step === LAST_STEP ? save() : go(step + 1))}
		onjump={go}
	>
		{#snippet status()}
			<!-- 어느 단계에서든 저장 (DESIGN 4.3절 "수정 모드") — 글자 하나 고치러 8단계를 넘기지 않게 -->
			<Button variant="outline" class="min-h-11 border-input" disabled={!canSave} onclick={save}>
				{saving ? '저장하는 중…' : '저장'}
			</Button>
		{/snippet}
		<RecordStepFields bind:form {step} {emotions} {distortions} {fields} />
	</StepShell>
{:else if loadError}
	<div class="space-y-4">
		<FormError message={loadError} />
		<Button variant="outline" class="min-h-11 border-input" onclick={load}>다시 시도</Button>
	</div>
{:else}
	<p class="text-muted-foreground">불러오는 중…</p>
{/if}

<AlertDialog.Root
	open={leaveTo !== null}
	onOpenChange={(o) => {
		if (!o) leaveTo = null;
	}}
>
	<AlertDialog.Content class="rounded-md">
		<AlertDialog.Header>
			<AlertDialog.Title class="font-serif">고친 내용을 저장하지 않았어요</AlertDialog.Title>
			<AlertDialog.Description>나가면 고친 내용은 사라져요.</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel class="min-h-11 border-input">계속 고치기</AlertDialog.Cancel>
			<Button variant="outline" class="min-h-11 border-input" onclick={leaveAnyway}
				>그냥 나가기</Button
			>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
