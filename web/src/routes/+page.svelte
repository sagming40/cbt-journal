<script lang="ts">
	import { onMount } from 'svelte';
	import { toggleMode } from 'mode-watcher';
	import CircleAlert from '@lucide/svelte/icons/circle-alert';
	import CircleCheck from '@lucide/svelte/icons/circle-check';
	import SunMoon from '@lucide/svelte/icons/sun-moon';
	import { api } from '#lib/api/client.js';
	import type { components } from '#lib/api/schema.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';

	type Health = components['schemas']['Health'];

	// $state: 값이 바뀌면 이 값을 쓰는 화면 조각만 다시 그린다.
	let health = $state<Health | null>(null);
	let unreachable = $state(false); // 서버에 아예 닿지 못함 (Go가 꺼짐 등)
	let loading = $state(true);

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
			// fetch 자체가 실패: 개발 서버까지 꺼졌거나 네트워크가 끊김
			health = null;
			unreachable = true;
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<main class="mx-auto max-w-md space-y-4 p-4">
	<header class="flex items-center justify-between">
		<h1 class="text-xl font-bold">CBT 저널</h1>
		<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="테마 바꾸기">
			<SunMoon />
		</Button>
	</header>

	<Card.Root>
		<Card.Header>
			<Card.Title>서버 상태</Card.Title>
			<Card.Description>브라우저 → Go → Postgres · Valkey 줄기 확인용 화면이야</Card.Description>
		</Card.Header>

		<Card.Content>
			{#if loading && !health && !unreachable}
				<p class="text-muted-foreground">확인하는 중…</p>
			{:else if unreachable}
				<!-- 오류는 빨강이 아니라 아이콘 + 굵은 글자 (DESIGN 6장 "오류 표시") -->
				<p class="flex items-center gap-2 font-semibold">
					<CircleAlert class="size-4" aria-hidden="true" /> 연결이 안 돼
				</p>
			{:else if health}
				<ul class="space-y-2">
					{#each [{ name: 'Postgres', status: health.postgres }, { name: 'Valkey', status: health.valkey }] as item (item.name)}
						<li class="flex items-center justify-between">
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
		</Card.Content>

		<Card.Footer>
			<Button onclick={load} disabled={loading}>다시 시도</Button>
		</Card.Footer>
	</Card.Root>
</main>
