<script lang="ts">
	// 로그인 · 가입 화면의 틀 (DESIGN 4.8절).
	// 데스크톱은 반으로 나눠 왼쪽 종이에 "이 앱이 뭔지"를 한 번에 보여 주고, 오른쪽에 폼. 모바일은 로고 + 폼만.
	import type { Snippet } from 'svelte';
	import Logo from '#lib/components/Logo.svelte';

	let { title, children }: { title: string; children: Snippet } = $props();

	// 쓴 내용을 로그에 남기지 않는 건 실제 규칙이다 (REQUIREMENTS 5.1절) — 지키지 않는 약속은 적지 않는다
	const points = [
		['10초 체크인', '지금 감정을 그 자리에서 붙잡아요'],
		['사고 기록', '떠오른 생각과 사실을 나눠 봐요'],
		['나만 보는 기록', '쓴 내용은 로그에도 남기지 않아요']
	];
</script>

<div class="min-h-dvh lg:grid lg:grid-cols-2">
	<aside
		class="relative hidden overflow-hidden border-r bg-card px-14 py-14 lg:flex lg:flex-col xl:px-20"
	>
		<Logo size={36} wordmark />
		<div class="relative z-10 my-auto">
			<p class="font-serif text-4xl leading-snug font-bold xl:text-[44px]">
				한 걸음 떨어져,<br />내 마음을 다시 보기
			</p>
			<ul class="mt-9 space-y-3.5">
				{#each points as [name, text] (name)}
					<li class="flex items-baseline gap-3">
						<span class="size-2 shrink-0 -translate-y-0.5 rounded-full bg-dot" aria-hidden="true"
						></span>
						<span>
							<b class="font-serif">{name}</b>
							<span class="text-muted-foreground">— {text}</span>
						</span>
					</li>
				{/each}
			</ul>
		</div>
		<!-- 꾸밈은 로고의 두 원을 크게 흐리게 깐 것 하나 (DESIGN 4.8절) -->
		<svg
			viewBox="0 0 64 64"
			class="pointer-events-none absolute top-10 -right-44 size-[560px] text-primary opacity-[0.07] dark:text-accent-text"
			aria-hidden="true"
		>
			<circle cx="22" cy="36" r="14" fill="currentColor" />
			<circle cx="43" cy="26" r="14" fill="none" stroke="currentColor" stroke-width="1.2" />
		</svg>
	</aside>

	<!-- 종이 위에 바로 쓴다 — 카드 상자 없이 (DESIGN 6장 "모양 D2") -->
	<main class="mx-auto flex min-h-dvh w-full max-w-sm flex-col justify-center px-6 py-10">
		<header class="mb-8">
			<div class="lg:hidden"><Logo size={44} wordmark tagline /></div>
			<h1 class="mt-8 font-serif text-2xl font-bold lg:mt-0 lg:text-3xl">{title}</h1>
		</header>
		{@render children()}
	</main>
</div>
