<script lang="ts">
	// 체크인 원 — S-02 가운데의 큰 원 버튼 (DESIGN 6장 "체크인 원", 4.1절).
	// "숨 한 번 고르고 누르기": 바깥 가는 테가 4초에 한 번 숨 쉬듯 커졌다 작아진다. OS "동작 줄이기"면 멈춘다 (DESIGN 8장).
	// 반짝이는 유리처럼 보이지 않게 납작한 잉크 원 + 옅은 테 두 겹.
	import { resolve } from '$app/paths';
</script>

<div class="flex flex-col items-center gap-3.5">
	<a
		href={resolve('/(app)/checkin')}
		class="orb relative flex size-[188px] flex-col items-center justify-center rounded-full bg-primary text-primary-foreground outline-offset-[30px]"
	>
		<span class="font-serif text-2xl font-bold">체크인</span>
		<span class="text-sm opacity-90">10초면 돼요</span>
	</a>
	<p class="text-sm text-muted-foreground">숨 한 번 고르고 눌러요</p>
</div>

<style>
	.orb {
		/* 가운데가 아주 조금 밝은 정도 — 토큰만으로 (primary-foreground = 흰색) */
		background-image: radial-gradient(
			circle at 50% 42%,
			color-mix(in oklab, var(--primary) 88%, var(--primary-foreground)),
			var(--primary) 70%
		);
		box-shadow:
			0 0 0 12px var(--accent-tint),
			0 0 0 26px color-mix(in oklab, var(--primary) 5%, transparent);
	}
	.orb::before {
		content: '';
		position: absolute;
		inset: -12px;
		border-radius: 9999px;
		border: 1px solid color-mix(in oklab, var(--primary) 35%, transparent);
		animation: breathe 4s ease-in-out infinite;
	}
	@keyframes breathe {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.9;
		}
		50% {
			transform: scale(1.09);
			opacity: 0.25;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.orb::before {
			animation: none;
		}
	}
</style>
