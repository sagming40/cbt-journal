<script lang="ts">
	import './layout.css';
	import { ModeWatcher } from 'mode-watcher';
	import favicon from '#lib/assets/favicon.svg';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>CBT 저널</title>
</svelte:head>

<!-- 시스템 · 사용자가 고른 테마를 보고 <html>에 dark 클래스를 붙였다 뗐다 한다 (DESIGN 6장 "테마") -->
<ModeWatcher />
<!-- "저장했어" 같은 짧은 알림 (DESIGN 5.1절 Sonner) -->
<Toaster position="top-center" />

{@render children()}
