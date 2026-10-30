<script lang="ts">
	// 로그인해야 쓰는 화면들의 틀. 내 정보를 먼저 확인하고, 확인되기 전에는 짧은 로딩만 보여 준다
	// (로그인 화면이 번쩍 보였다 사라지지 않게, DESIGN 3.1절). 401이면 api 클라이언트가 로그인 화면으로 보낸다.
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import House from '@lucide/svelte/icons/house';
	import Settings from '@lucide/svelte/icons/settings';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR } from '#lib/api/errors.js';
	import { auth } from '#lib/auth.svelte.js';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
	let offline = $state(false);

	async function loadMe() {
		offline = false;
		try {
			const { data } = await api.GET('/api/me');
			if (data) auth.set(data);
		} catch {
			offline = true; // 서버에 아예 닿지 못함
		}
	}

	onMount(() => {
		if (!auth.user) loadMe();
	});

	// 메뉴 (DESIGN 3.2절). 기록 · 통계는 그 화면을 만드는 단계(M6)에서 더한다.
	const nav = [
		{ href: resolve('/(app)'), label: '오늘', icon: House },
		{ href: resolve('/(app)/settings'), label: '설정', icon: Settings }
	];
	const isActive = (href: string) => page.url.pathname === href;
</script>

{#if auth.user}
	<div class="min-h-dvh md:flex">
		<!-- 데스크톱: 왼쪽 사이드바 / 모바일: 아래 탭 (기준 48rem = md, DESIGN 3.2절) -->
		<nav
			aria-label="메뉴"
			class="fixed inset-x-0 bottom-0 z-10 flex border-t bg-background md:static md:w-56 md:flex-col md:gap-1 md:border-t-0 md:border-r md:p-3"
		>
			<p class="hidden px-3 py-2 text-lg font-bold md:block">CBT 저널</p>
			{#each nav as item (item.href)}
				<a
					href={item.href}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class={[
						'flex min-h-11 flex-1 flex-col items-center justify-center gap-0.5 text-xs md:flex-none md:flex-row md:justify-start md:gap-2 md:rounded-lg md:px-3 md:text-sm',
						isActive(item.href)
							? 'font-semibold text-accent-text md:bg-accent-tint'
							: 'text-muted-foreground hover:text-foreground'
					]}
				>
					<item.icon class="size-5 md:size-4" aria-hidden="true" />
					{item.label}
				</a>
			{/each}
		</nav>

		<!-- 본문은 최대 40rem — 글을 쓰는 앱이라 한 줄이 너무 길면 읽기 · 쓰기가 불편하다 (DESIGN 3.2절) -->
		<main class="mx-auto w-full max-w-[40rem] p-4 pb-24 md:pb-4">
			{@render children()}
		</main>
	</div>
{:else if offline}
	<div class="mx-auto flex min-h-dvh max-w-md flex-col items-center justify-center gap-4 p-4">
		<FormError message={NETWORK_ERROR} />
		<Button onclick={loadMe}>다시 시도</Button>
	</div>
{:else}
	<p class="flex min-h-dvh items-center justify-center text-muted-foreground">불러오는 중…</p>
{/if}
