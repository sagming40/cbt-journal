<script lang="ts">
	// S-01 로그인 (DESIGN 4.8절).
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors, retryAfterText } from '#lib/api/errors.js';
	import { auth } from '#lib/auth.svelte.js';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Checkbox } from '#lib/components/ui/checkbox/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { safeNext } from '#lib/safe-next.js';

	let email = $state('');
	let password = $state('');
	// 로그인 유지는 기본 해제 — 회사 PC처럼 자리를 비우는 기기에서 실수로 30일 유지되지 않게 (REQUIREMENTS A1).
	let remember = $state(false);
	let submitting = $state(false);
	let message = $state('');
	let fields = $state<Record<string, string>>({});
	let allowSignup = $state(false);

	// 로그인 뒤 갈 곳. 바깥 주소는 무시하고 / (DESIGN 3.1절).
	const next = $derived(safeNext(page.url.searchParams.get('next')));

	function goNext() {
		// next는 safeNext가 앱 안 경로만 남긴 값이다.
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- 동적 경로라 resolve(라우트 id)를 쓸 수 없고, safeNext로 검증했다
		return goto(next, { replace: true });
	}

	onMount(async () => {
		try {
			// 이미 로그인돼 있으면 바로 넘어간다. 401은 api 클라이언트가 이 화면에서는 그냥 둔다.
			const [me, config] = await Promise.all([api.GET('/api/me'), api.GET('/api/config')]);
			allowSignup = config.data?.allow_signup ?? false;
			if (me.data) {
				auth.set(me.data);
				await goNext();
			}
		} catch {
			// 서버가 꺼져 있으면 제출할 때 알려 준다.
		}
	});

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		submitting = true;
		message = '';
		fields = {};
		try {
			const { data, error, response } = await api.POST('/api/auth/login', {
				body: { email, password, remember }
			});
			if (data) {
				auth.set(data);
				await goNext();
				return;
			}
			if (response.status === 429) {
				const wait = retryAfterText(response);
				message = wait ? `잠시 뒤에 다시 시도해 주세요 (${wait})` : '잠시 뒤에 다시 시도해 주세요';
			} else {
				fields = fieldErrors(error);
				message = Object.keys(fields).length > 0 ? '' : errorMessage(error);
			}
		} catch {
			message = NETWORK_ERROR;
		} finally {
			submitting = false;
		}
	}
</script>

<main class="mx-auto flex min-h-dvh max-w-sm flex-col justify-center p-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-xl">CBT 저널</Card.Title>
			<Card.Description>로그인하고 이어서 기록해 보세요</Card.Description>
		</Card.Header>
		<Card.Content>
			<form class="space-y-4" onsubmit={submit} novalidate>
				<div class="space-y-1.5">
					<Label for="email">이메일</Label>
					<Input
						id="email"
						type="email"
						autocomplete="email"
						bind:value={email}
						aria-invalid={fields.email ? 'true' : undefined}
						aria-describedby={fields.email ? 'email-error' : undefined}
						required
					/>
					<FormError id="email-error" message={fields.email} />
				</div>
				<div class="space-y-1.5">
					<Label for="password">비밀번호</Label>
					<Input
						id="password"
						type="password"
						autocomplete="current-password"
						bind:value={password}
						aria-invalid={fields.password ? 'true' : undefined}
						aria-describedby={fields.password ? 'password-error' : undefined}
						required
					/>
					<FormError id="password-error" message={fields.password} />
				</div>
				<div class="space-y-1">
					<div class="flex min-h-11 items-center gap-2">
						<Checkbox id="remember" bind:checked={remember} />
						<Label for="remember">로그인 유지</Label>
					</div>
					<p class="text-xs text-muted-foreground">공용 · 회사 PC에서는 체크하지 마세요</p>
				</div>
				<FormError {message} />
				<Button type="submit" class="min-h-11 w-full" disabled={submitting}>
					{submitting ? '로그인하는 중…' : '로그인'}
				</Button>
			</form>
		</Card.Content>
		{#if allowSignup}
			<Card.Footer class="justify-center text-sm">
				<a href={resolve('/signup')} class="text-accent-text underline-offset-4 hover:underline"
					>처음이세요? 가입하기</a
				>
			</Card.Footer>
		{/if}
	</Card.Root>
</main>
