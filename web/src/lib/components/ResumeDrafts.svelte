<script lang="ts">
	// S-02 "이어 쓰기" (DESIGN 4.1절): 초안이 있을 때만, 최근 저장순으로. 어느 기기에서 쓴 초안이든 여기서 이어 쓴다.
	// 초안 저장소(Valkey)가 죽어 있어도 오늘 화면은 멀쩡해야 한다 — 이 구역만 한 줄로 알린다 (decisions 0005).
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import Trash from '@lucide/svelte/icons/trash-2';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import FormError from '#lib/components/FormError.svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { LAST_STEP } from '#lib/record-form.js';
	import { formatAgo } from '#lib/tz.js';

	type DraftSummary = components['schemas']['DraftSummary'];

	let drafts = $state<DraftSummary[]>([]);
	let unavailable = $state('');
	let deleteMessage = $state('');
	let target = $state<DraftSummary | null>(null);
	const now = new Date();

	async function load() {
		try {
			const { data, error } = await api.GET('/api/drafts');
			if (data) {
				drafts = data.items;
				unavailable = '';
			} else {
				unavailable = errorMessage(error);
			}
		} catch {
			unavailable = NETWORK_ERROR;
		}
	}

	onMount(load);

	async function remove() {
		const d = target;
		target = null;
		if (!d) return;
		deleteMessage = '';
		try {
			const { error, response } = await api.DELETE('/api/drafts/{id}', {
				params: { path: { id: d.id } }
			});
			// 이미 없어도(404) 사용자가 원한 결과는 같다
			if (response.status === 204 || response.status === 404) {
				drafts = drafts.filter((x) => x.id !== d.id);
				toast.success('지웠어요');
			} else {
				deleteMessage = errorMessage(error);
			}
		} catch {
			deleteMessage = NETWORK_ERROR;
		}
	}
</script>

{#if drafts.length > 0}
	<section class="mt-8 border-t pt-4" aria-labelledby="resume-title">
		<h2 id="resume-title" class="font-serif text-base font-bold">이어 쓰기</h2>
		<ul class="mt-1">
			{#each drafts as d (d.id)}
				<li class="flex items-center gap-1 border-b border-dashed last:border-b-0">
					<!-- eslint-disable svelte/no-navigation-without-resolve -- resolve로 만든 주소에 초안 id만 붙였다 -->
					<a
						href={`${resolve('/(app)/records/new')}?draft=${d.id}`}
						class="min-w-0 flex-1 py-3 hover:bg-muted/40"
					>
						<span class="block truncate">{d.situation_preview || '아직 상황을 안 썼어요'}</span>
						<span class="block text-xs text-muted-foreground tabular-nums">
							{d.step}/{LAST_STEP} · {formatAgo(d.updated_at, now)}
						</span>
					</a>
					<!-- eslint-enable svelte/no-navigation-without-resolve -->
					<button
						type="button"
						class="flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
						aria-label={`초안 지우기: ${d.situation_preview}`}
						onclick={() => (target = d)}
					>
						<Trash class="size-4" aria-hidden="true" />
					</button>
				</li>
			{/each}
		</ul>
		<FormError message={deleteMessage} />
	</section>
{:else if unavailable}
	<p class="mt-6 text-center text-xs text-muted-foreground">
		이어 쓸 초안을 지금은 불러올 수 없어요
	</p>
{/if}

<AlertDialog.Root
	open={target !== null}
	onOpenChange={(o) => {
		if (!o) target = null;
	}}
>
	<AlertDialog.Content class="rounded-md">
		<AlertDialog.Header>
			<AlertDialog.Title class="font-serif">이 초안을 지울까요?</AlertDialog.Title>
			<AlertDialog.Description>지운 초안은 되돌릴 수 없어요.</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel class="min-h-11 border-input">그대로 두기</AlertDialog.Cancel>
			<!-- AlertDialog.Action이 아니라 Button: Action은 먼저 창을 닫아 target이 비워진 뒤에 onclick이 돈다 -->
			<Button class="min-h-11" onclick={remove}>지우기</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
