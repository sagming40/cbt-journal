<script lang="ts">
	// 로그인해야 쓰는 화면들의 틀. 내 정보를 먼저 확인하고, 확인되기 전에는 짧은 로딩만 보여 준다
	// (로그인 화면이 번쩍 보였다 사라지지 않게, DESIGN 3.1절). 401이면 api 클라이언트가 로그인 화면으로 보낸다.
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import House from '@lucide/svelte/icons/house';
	import Plus from '@lucide/svelte/icons/plus';
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
	// 체크인 · 사고 기록을 쓰는 중에는 메뉴를 숨긴다: 화면 아래 고정 버튼 자리이고,
	// 실수로 다른 탭을 눌러 흐름이 끊기지 않게 (DESIGN 3.2절). 사고 기록 상세(/records/:id)는 읽는 화면이라 메뉴가 있다.
	const focusMode = $derived(
		page.url.pathname.startsWith(resolve('/(app)/checkin')) ||
			page.url.pathname === resolve('/(app)/records/new') ||
			page.url.pathname.endsWith('/edit')
	);
</script>

{#if auth.user}
	<div class="min-h-dvh md:flex">
		<!-- 데스크톱: 왼쪽 사이드바 (기준 48rem = md, DESIGN 3.2절). 스크롤해도 제자리에 붙어 있다 -->
		<aside
			class="sticky top-0 hidden h-dvh w-60 shrink-0 flex-col border-r bg-card px-4 pt-7 pb-5 md:flex"
		>
			<a href={resolve('/(app)')} class="px-1" aria-label="CBT 저널 — 오늘">
				<Logo size={34} wordmark tagline />
			</a>
			<nav aria-label="메뉴" class="mt-8 flex flex-col gap-0.5">
				{#each nav as item (item.href)}
					<a
						href={item.href}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class={[
							'flex min-h-11 items-center gap-2.5 rounded-md px-3 text-[15px]',
							isActive(item.href) ? 'bg-accent-tint font-bold text-accent-text' : 'hover:bg-muted'
						]}
					>
						<item.icon class="size-[18px]" aria-hidden="true" />
						{item.label}
					</a>
				{/each}
			</nav>
			<!-- 가장 자주 하는 동작은 메뉴 사이에 묻지 않고 맨 아래 따로 (DESIGN 3.2절) -->
			<Button href={resolve('/(app)/checkin')} class="mt-auto min-h-11 w-full font-semibold">
				<Plus class="size-4" aria-hidden="true" />체크인
			</Button>
			<div class="mt-4 flex items-center gap-2.5 border-t border-dashed px-1 pt-4 text-sm">
				<span
					class="flex size-8 shrink-0 items-center justify-center rounded-full border border-input text-[13px] font-bold"
					aria-hidden="true">{auth.user.nickname.slice(0, 1)}</span
				>
				<span class="min-w-0 leading-tight">
					<span class="block truncate">{auth.user.nickname}</span>
					<span class="block truncate text-xs text-muted-foreground">{auth.user.timezone}</span>
				</span>
			</div>
		</aside>

		<!-- 모바일: 아래 탭. 체크인 중에는 숨긴다(그 자리에 고정 [저장], DESIGN 3.2절) -->
		<nav
			aria-label="메뉴"
			class={[
				'fixed inset-x-0 bottom-0 z-10 border-t bg-background pb-[env(safe-area-inset-bottom)] md:hidden',
				focusMode ? 'hidden' : 'flex'
			]}
		>
			{#each nav as item (item.href)}
				<a
					href={item.href}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class={[
						'flex min-h-14 flex-1 flex-col items-center justify-center gap-1 text-xs',
						isActive(item.href)
							? 'font-semibold text-accent-text'
							: 'text-muted-foreground hover:text-foreground'
					]}
				>
					<item.icon class="size-5" aria-hidden="true" />
					<!-- 고른 탭은 글자 아래 밑줄 2px (DESIGN 6장 "모양 D2") -->
					<span
						class={[
							'border-b-2 pb-0.5',
							isActive(item.href) ? 'border-primary dark:border-accent-text' : 'border-transparent'
						]}>{item.label}</span
					>
				</a>
			{/each}
		</nav>

		<!-- 틀은 최대 68rem, 화면마다 2단으로 나눈다. 글줄 한 줄은 각 화면에서 40rem 안으로 (DESIGN 3.2절) -->
		<main
			class={[
				'mx-auto w-full max-w-[68rem] min-w-0 px-5 pt-6 md:px-12 md:pt-11 md:pb-16',
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
