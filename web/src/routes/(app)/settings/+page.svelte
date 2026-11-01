<script lang="ts">
	// S-08 설정 — M2에서는 계정 · 표시 묶음과 로그아웃 (DESIGN 4.7절). 리마인더 · 데이터 묶음은 M7 · M8.
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { setMode, userPrefersMode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors, retryAfterText } from '#lib/api/errors.js';
	import { auth } from '#lib/auth.svelte.js';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Separator } from '#lib/components/ui/separator/index.js';
	import * as ToggleGroup from '#lib/components/ui/toggle-group/index.js';

	// 이 화면은 (app) 틀 안이라 auth.user가 있을 때만 그려진다.
	const me = $derived(auth.user!);

	// ── 닉네임 ──
	let nickname = $state(auth.user?.nickname ?? '');
	let nicknameError = $state('');
	let savingNickname = $state(false);

	async function saveNickname(event: SubmitEvent) {
		event.preventDefault();
		savingNickname = true;
		nicknameError = '';
		try {
			const { data, error } = await api.PATCH('/api/me', { body: { nickname } });
			if (data) {
				auth.set(data);
				nickname = data.nickname;
				toast.success('저장했어요');
			} else {
				nicknameError = fieldErrors(error).nickname ?? errorMessage(error);
			}
		} catch {
			nicknameError = NETWORK_ERROR;
		} finally {
			savingNickname = false;
		}
	}

	// ── 시간대 ── "하루"의 경계를 정한다 (REQUIREMENTS A4). 휴대폰에서는 기본 선택창이 목록을 고르기 편해서 select를 쓴다.
	const zones: string[] = (() => {
		const list = Intl.supportedValuesOf('timeZone');
		const current = auth.user?.timezone;
		return current && !list.includes(current) ? [current, ...list] : list;
	})();
	let timezoneError = $state('');

	async function saveTimezone(event: Event) {
		const timezone = (event.currentTarget as HTMLSelectElement).value;
		timezoneError = '';
		try {
			const { data, error } = await api.PATCH('/api/me', { body: { timezone } });
			if (data) {
				auth.set(data);
				toast.success('시간대를 바꿨어요');
			} else {
				timezoneError = fieldErrors(error).timezone ?? errorMessage(error);
			}
		} catch {
			timezoneError = NETWORK_ERROR;
		}
	}

	// ── 비밀번호 변경 (A3) ──
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let pwFields = $state<Record<string, string>>({});
	let pwMessage = $state('');
	let changingPassword = $state(false);

	async function changePassword(event: SubmitEvent) {
		event.preventDefault();
		pwFields = {};
		pwMessage = '';
		if (newPassword !== confirmPassword) {
			pwFields = { confirm: '두 비밀번호가 달라요' };
			return;
		}
		changingPassword = true;
		try {
			const { error, response } = await api.PUT('/api/me/password', {
				body: { current_password: currentPassword, new_password: newPassword }
			});
			if (response.status === 204) {
				currentPassword = newPassword = confirmPassword = '';
				toast.success('비밀번호를 바꿨어요. 다른 기기는 로그아웃됐어요');
			} else if (response.status === 429) {
				const wait = retryAfterText(response);
				pwMessage = wait
					? `잠시 뒤에 다시 시도해 주세요 (${wait})`
					: '잠시 뒤에 다시 시도해 주세요';
			} else {
				pwFields = fieldErrors(error);
				pwMessage = Object.keys(pwFields).length > 0 ? '' : errorMessage(error);
			}
		} catch {
			pwMessage = NETWORK_ERROR;
		} finally {
			changingPassword = false;
		}
	}

	// ── 다른 기기 모두 로그아웃 (A8) ──
	let revoking = $state(false);
	let revokeMessage = $state('');

	async function logoutOthers() {
		revoking = true;
		revokeMessage = '';
		try {
			const { data, error } = await api.POST('/api/me/sessions/logout-others');
			if (data) {
				toast.success(
					data.revoked > 0
						? `다른 기기 ${data.revoked}곳을 로그아웃했어요`
						: '다른 기기에 로그인된 곳이 없어요'
				);
			} else {
				revokeMessage = errorMessage(error);
			}
		} catch {
			revokeMessage = NETWORK_ERROR;
		} finally {
			revoking = false;
		}
	}

	// ── 로그아웃 ──
	let logoutMessage = $state('');

	async function logout() {
		logoutMessage = '';
		try {
			await api.POST('/api/auth/logout');
			auth.clear();
			await goto(resolve('/login'), { replace: true });
		} catch {
			logoutMessage = NETWORK_ERROR;
		}
	}

	// ── 데스크톱 목차 (DESIGN 4.7절): 지금 화면 위쪽에 걸린 묶음에 선을 긋는다 ──
	const toc = [
		{ id: 'account', label: '계정' },
		{ id: 'display', label: '표시' }
	];
	let current = $state('account');

	onMount(() => {
		// 화면 위 40% 안에 들어온 묶음을 "지금 보는 묶음"으로 친다
		const io = new IntersectionObserver(
			(entries) => {
				for (const e of entries) if (e.isIntersecting) current = e.target.id;
			},
			{ rootMargin: '0px 0px -60% 0px' }
		);
		for (const { id } of toc) {
			const el = document.getElementById(id);
			if (el) io.observe(el);
		}
		return () => io.disconnect();
	});
</script>

<h1 class="mb-6 font-serif text-2xl font-bold md:mb-10 md:border-b md:pb-6 md:text-[34px]">설정</h1>

<!-- 데스크톱(64rem 이상): 목차 | 내용 (DESIGN 4.7절). 모바일은 목차 없이 선으로 나눈 구역을 위아래로 (6장 "모양 D2") -->
<div class="lg:grid lg:grid-cols-[9rem_1fr] lg:items-start lg:gap-10">
	<nav aria-label="설정 목차" class="sticky top-8 hidden flex-col lg:flex">
		{#each toc as item (item.id)}
			<a
				href={`#${item.id}`}
				onclick={() => (current = item.id)}
				aria-current={current === item.id ? 'location' : undefined}
				class={[
					'border-l-2 px-3 py-2',
					current === item.id
						? 'border-primary font-bold text-accent-text dark:border-accent-text'
						: 'border-border text-muted-foreground hover:text-foreground'
				]}>{item.label}</a
			>
		{/each}
	</nav>

	<div class="space-y-10 md:space-y-5">
		<section
			id="account"
			class="scroll-mt-8 border-t pt-5 md:rounded-lg md:border md:bg-card md:px-7 md:pt-6 md:pb-7"
			aria-labelledby="account-title"
		>
			<h2 id="account-title" class="mb-5 font-serif text-lg font-bold">계정</h2>
			<div class="space-y-6 md:max-w-md">
				<form class="space-y-1.5" onsubmit={saveNickname} novalidate>
					<Label for="nickname">닉네임</Label>
					<div class="flex gap-2">
						<Input
							id="nickname"
							bind:value={nickname}
							autocomplete="nickname"
							aria-invalid={nicknameError ? 'true' : undefined}
						/>
						<Button
							type="submit"
							variant="outline"
							class="min-h-11 border-input"
							disabled={savingNickname || nickname === me.nickname}>저장</Button
						>
					</div>
					<FormError message={nicknameError} />
				</form>

				<div class="space-y-1.5">
					<p class="text-sm font-medium">이메일</p>
					<p class="text-sm text-muted-foreground">{me.email}</p>
				</div>

				<Separator />

				<form class="space-y-3" onsubmit={changePassword} novalidate>
					<p class="text-sm font-medium">비밀번호 바꾸기</p>
					<div class="space-y-1.5">
						<Label for="current-password">지금 비밀번호</Label>
						<Input
							id="current-password"
							type="password"
							autocomplete="current-password"
							bind:value={currentPassword}
							aria-invalid={pwFields.current_password ? 'true' : undefined}
						/>
						<FormError message={pwFields.current_password} />
					</div>
					<div class="space-y-1.5">
						<Label for="new-password">새 비밀번호</Label>
						<Input
							id="new-password"
							type="password"
							autocomplete="new-password"
							bind:value={newPassword}
							aria-invalid={pwFields.new_password ? 'true' : undefined}
						/>
						{#if !pwFields.new_password}<p class="text-xs text-muted-foreground">
								12자 이상이에요. 바꾸면 다른 기기는 로그아웃돼요
							</p>{/if}
						<FormError message={pwFields.new_password} />
					</div>
					<div class="space-y-1.5">
						<Label for="confirm-password">새 비밀번호 한 번 더</Label>
						<Input
							id="confirm-password"
							type="password"
							autocomplete="new-password"
							bind:value={confirmPassword}
							aria-invalid={pwFields.confirm ? 'true' : undefined}
						/>
						<FormError message={pwFields.confirm} />
					</div>
					<FormError message={pwMessage} />
					<Button
						type="submit"
						variant="outline"
						class="min-h-11 border-input"
						disabled={changingPassword || !currentPassword || !newPassword}
					>
						{changingPassword ? '바꾸는 중…' : '비밀번호 바꾸기'}
					</Button>
				</form>

				<Separator />

				<div class="space-y-1.5">
					<p class="text-sm font-medium">다른 기기 모두 로그아웃</p>
					<p class="text-xs text-muted-foreground">이 기기만 남기고 로그인된 곳을 모두 끊어요</p>
					<Button
						variant="outline"
						class="min-h-11 border-input"
						onclick={logoutOthers}
						disabled={revoking}
					>
						{revoking ? '끊는 중…' : '다른 기기 로그아웃'}
					</Button>
					<FormError message={revokeMessage} />
				</div>
			</div>
		</section>

		<section
			id="display"
			class="scroll-mt-8 border-t pt-5 md:rounded-lg md:border md:bg-card md:px-7 md:pt-6 md:pb-7"
			aria-labelledby="display-title"
		>
			<h2 id="display-title" class="mb-5 font-serif text-lg font-bold">표시</h2>
			<div class="space-y-6 md:max-w-md">
				<div class="space-y-1.5">
					<Label for="timezone">시간대</Label>
					<select
						id="timezone"
						class="h-11 w-full rounded-md border border-input bg-card px-3 text-sm"
						value={me.timezone}
						onchange={saveTimezone}
					>
						{#each zones as zone (zone)}
							<option value={zone}>{zone}</option>
						{/each}
					</select>
					<p class="text-xs text-muted-foreground">"오늘"의 시작과 끝을 이 시간대로 정해요</p>
					<FormError message={timezoneError} />
				</div>

				<div class="space-y-1.5">
					<p class="text-sm font-medium" id="theme-label">테마</p>
					<!-- 테마는 기기마다 고른다(이 브라우저에 저장). 기본은 시스템 설정을 따른다 (DESIGN 6장) -->
					<ToggleGroup.Root
						type="single"
						variant="outline"
						aria-labelledby="theme-label"
						value={userPrefersMode.current}
						onValueChange={(v) => v && setMode(v as 'system' | 'light' | 'dark')}
					>
						<ToggleGroup.Item value="system" class="min-h-11 px-4">시스템</ToggleGroup.Item>
						<ToggleGroup.Item value="light" class="min-h-11 px-4">라이트</ToggleGroup.Item>
						<ToggleGroup.Item value="dark" class="min-h-11 px-4">다크</ToggleGroup.Item>
					</ToggleGroup.Root>
				</div>
			</div>
		</section>

		<div class="space-y-1.5 border-t pt-6 md:border-t-0 md:pt-2">
			<Button
				variant="outline"
				class="min-h-11 w-full border-input md:w-auto md:px-8"
				onclick={logout}>로그아웃</Button
			>
			<FormError message={logoutMessage} />
		</div>
	</div>
</div>
