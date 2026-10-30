<script lang="ts">
	// 가입 (REQUIREMENTS A2). ALLOW_SIGNUP이 꺼져 있으면 서버가 403을 주므로 안내만 보여 준다.
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors, retryAfterText } from '#lib/api/errors.js';
	import { auth } from '#lib/auth.svelte.js';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';

	let email = $state('');
	let nickname = $state('');
	let password = $state('');
	let confirm = $state('');
	let submitting = $state(false);
	let message = $state('');
	let fields = $state<Record<string, string>>({});
	let allowSignup = $state<boolean | null>(null);

	onMount(async () => {
		try {
			const { data } = await api.GET('/api/config');
			allowSignup = data?.allow_signup ?? false;
		} catch {
			allowSignup = false;
			message = NETWORK_ERROR;
		}
	});

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		message = '';
		fields = {};
		// 비밀번호 확인은 화면에서만 본다 — 오타로 모르는 비밀번호가 되지 않게.
		if (password !== confirm) {
			fields = { confirm: '두 비밀번호가 달라요' };
			return;
		}
		submitting = true;
		try {
			const { data, error, response } = await api.POST('/api/auth/signup', {
				body: { email, nickname, password }
			});
			if (data) {
				auth.set(data);
				await goto(resolve('/(app)'), { replace: true });
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
			<Card.Title class="text-xl">가입하기</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if allowSignup === false}
				<FormError message={message || '지금은 가입을 받지 않아요'} />
			{:else if allowSignup}
				<form class="space-y-4" onsubmit={submit} novalidate>
					<div class="space-y-1.5">
						<Label for="email">이메일</Label>
						<Input
							id="email"
							type="email"
							autocomplete="email"
							bind:value={email}
							aria-invalid={fields.email ? 'true' : undefined}
						/>
						<FormError message={fields.email} />
					</div>
					<div class="space-y-1.5">
						<Label for="nickname">닉네임</Label>
						<Input
							id="nickname"
							autocomplete="nickname"
							bind:value={nickname}
							aria-invalid={fields.nickname ? 'true' : undefined}
						/>
						{#if !fields.nickname}<p class="text-xs text-muted-foreground">30자까지</p>{/if}
						<FormError message={fields.nickname} />
					</div>
					<div class="space-y-1.5">
						<Label for="password">비밀번호</Label>
						<Input
							id="password"
							type="password"
							autocomplete="new-password"
							bind:value={password}
							aria-invalid={fields.password ? 'true' : undefined}
						/>
						{#if !fields.password}<p class="text-xs text-muted-foreground">
								12자 이상이에요. 문장처럼 길게 써도 돼요
							</p>{/if}
						<FormError message={fields.password} />
					</div>
					<div class="space-y-1.5">
						<Label for="confirm">비밀번호 한 번 더</Label>
						<Input
							id="confirm"
							type="password"
							autocomplete="new-password"
							bind:value={confirm}
							aria-invalid={fields.confirm ? 'true' : undefined}
						/>
						<FormError message={fields.confirm} />
					</div>
					<FormError {message} />
					<Button type="submit" class="min-h-11 w-full" disabled={submitting}>
						{submitting ? '가입하는 중…' : '가입'}
					</Button>
				</form>
			{/if}
		</Card.Content>
		<Card.Footer class="justify-center text-sm">
			<a href={resolve('/login')} class="text-accent-text underline-offset-4 hover:underline"
				>로그인으로</a
			>
		</Card.Footer>
	</Card.Root>
</main>
