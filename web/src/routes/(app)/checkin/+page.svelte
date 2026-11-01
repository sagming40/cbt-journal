<script lang="ts">
	// S-03 체크인 작성 (DESIGN 4.2절, REQUIREMENTS C1). 저장하면 오늘로 돌아가 "기록했어요".
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import X from '@lucide/svelte/icons/x';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import CheckinForm from '#lib/components/CheckinForm.svelte';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';

	type Emotion = components['schemas']['Emotion'];
	type CheckinInput = components['schemas']['CheckinInput'];

	let emotions = $state<Emotion[] | null>(null);
	let loadError = $state('');
	let submitting = $state(false);
	let fields = $state<Record<string, string>>({});
	let message = $state('');

	async function loadEmotions() {
		loadError = '';
		try {
			const { data, error } = await api.GET('/api/emotions');
			if (data) emotions = data.items;
			else loadError = errorMessage(error);
		} catch {
			loadError = NETWORK_ERROR;
		}
	}

	onMount(loadEmotions);

	async function save(body: CheckinInput) {
		submitting = true;
		fields = {};
		message = '';
		try {
			const { data, error } = await api.POST('/api/checkins', { body });
			if (data) {
				toast.success('기록했어요');
				await goto(resolve('/(app)'));
				return;
			}
			fields = fieldErrors(error);
			message = Object.keys(fields).length > 0 ? '' : errorMessage(error);
		} catch {
			// 연결이 끊겨도 고른 감정은 화면에 그대로 남아 있다 — 다시 누르면 된다
			message = NETWORK_ERROR;
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>체크인 · CBT 저널</title></svelte:head>

<header class="mb-6 flex items-center gap-2 md:mb-9 md:border-b md:pb-4">
	<a
		href={resolve('/(app)')}
		class="-ml-2 flex size-11 items-center justify-center rounded-md hover:bg-muted"
		aria-label="닫기"
	>
		<X class="size-5" aria-hidden="true" />
	</a>
	<h1 class="font-serif text-lg font-bold">체크인</h1>
</header>

{#if emotions}
	<CheckinForm {emotions} {submitting} {fields} {message} onsubmit={save} />
{:else if loadError}
	<div class="space-y-4">
		<FormError message={loadError} />
		<Button variant="outline" class="min-h-11 border-input" onclick={loadEmotions}>다시 시도</Button
		>
	</div>
{:else}
	<p class="text-muted-foreground">불러오는 중…</p>
{/if}
