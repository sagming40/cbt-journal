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
	import Logo from '#lib/components/Logo.svelte';
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
	// 체크인 쓰는 중에는 메뉴를 숨긴다: 화면 아래 고정 [저장] 자리이고, 실수로 다른 탭을 눌러 흐름이 끊기지 않게
	// (DESIGN 3.2절 — S-04와 같은 이유)
	const focusMode = $derived(page.url.pathname.startsWith(resolve('/(app)/checkin')));
</script>

{#if auth.user}
	<div class="min-h-dvh md:flex">
		<!-- 데스크톱: 왼쪽 사이드바 / 모바일: 아래 탭 (기준 48rem = md, DESIGN 3.2절) -->
		<nav
			aria-label="메뉴"
			class={[
				'fixed inset-x-0 bottom-0 z-10 border-t bg-background pb-[env(safe-area-inset-bottom)] md:static md:flex md:min-h-dvh md:w-56 md:flex-col md:gap-1 md:border-t-0 md:border-r md:p-4 md:pb-4',
				// 체크인 중에는 모바일 하단 탭만 숨긴다(데스크톱 사이드바는 화면을 가리지 않아 둔다)
				focusMode ? 'hidden' : 'flex'
			]}
		>
			<a
				href={resolve('/(app)')}
				class="mb-6 hidden px-2 pt-1 md:block"
				aria-label="CBT 저널 — 오늘"
			>
				<Logo size={30} wordmark />
			</a>
			{#each nav as item (item.href)}
				<a
					href={item.href}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class={[
						'flex min-h-14 flex-1 flex-col items-center justify-center gap-1 text-xs md:min-h-11 md:flex-none md:flex-row md:justify-start md:gap-2.5 md:rounded-md md:px-3 md:text-sm',
						isActive(item.href)
							? 'font-semibold text-accent-text md:bg-accent-tint'
							: 'text-muted-foreground hover:text-foreground'
					]}
				>
					<item.icon class="size-5 md:size-4" aria-hidden="true" />
					<!-- 모바일 하단 탭: 고른 탭은 글자 아래 밑줄 2px (DESIGN 6장 "모양 D2"). 데스크톱은 옅은 바탕으로 -->
					<span
						class={[
							'border-b-2 pb-0.5 md:border-b-0 md:pb-0',
							isActive(item.href) ? 'border-primary dark:border-accent-text' : 'border-transparent'
						]}>{item.label}</span
					>
				</a>
			{/each}
		</nav>

		<!-- 본문은 최대 40rem — 글을 쓰는 앱이라 한 줄이 너무 길면 읽기 · 쓰기가 불편하다 (DESIGN 3.2절) -->
		<main
			class={[
				'mx-auto w-full max-w-[40rem] px-5 pt-6 md:px-8 md:pt-10 md:pb-10',
				focusMode ? 'pb-6' : 'pb-28'
			]}
		>
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
