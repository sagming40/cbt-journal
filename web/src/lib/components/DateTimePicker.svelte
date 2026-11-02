<script lang="ts">
	// 시각 고르기 (DESIGN 5.2절). 칸 전체가 버튼이다 — 브라우저 기본 datetime-local은 달력 아이콘만 눌러야 열려서 바꿨다.
	// 값은 datetime-local과 같은 "YYYY-MM-DDTHH:mm"(사용자 시간대 벽시계)라, 실제 시각으로 바꾸는 일은 tz.ts가 그대로 맡는다.
	// 시 · 분도 칸 어디를 눌러도 열리도록 기본 select를 쓴다. 휴대폰에서는 기기 선택창이 떠서 고르기도 편하다.
	import { parseDate } from '@internationalized/date';
	import CalendarDays from '@lucide/svelte/icons/calendar-days';
	import { Calendar } from '#lib/components/ui/calendar/index.js';
	import * as Popover from '#lib/components/ui/popover/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { formatLocalInput, toLocalInput } from '#lib/tz.js';

	let {
		id,
		value = $bindable(),
		tz,
		invalid = false,
		labelledby
	}: {
		id?: string;
		value: string;
		tz: string;
		invalid?: boolean;
		labelledby?: string;
	} = $props();

	let open = $state(false);
	const pad = (n: number) => String(n).padStart(2, '0');
	const HOURS = Array.from({ length: 24 }, (_, i) => pad(i));
	const MINUTES = Array.from({ length: 60 }, (_, i) => pad(i));

	const date = $derived(value.slice(0, 10));
	const hour = $derived(value.slice(11, 13));
	const minute = $derived(value.slice(14, 16));
	// 미래 기록은 받지 않는다(DATA_MODEL 3.4절). 오늘 안의 미래 시각은 저장할 때 서버가 알려 준다.
	const today = $derived(parseDate(toLocalInput(new Date(), tz).slice(0, 10)));
	// 달력이 지금 보여 주는 달. 열 때마다 고른 날짜의 달로 돌아온다.
	let shown = $state(parseDate(value.slice(0, 10)));
</script>

<Popover.Root
	bind:open
	onOpenChange={(o) => {
		if (o) shown = parseDate(date);
	}}
>
	<Popover.Trigger
		{id}
		aria-labelledby={labelledby}
		aria-invalid={invalid ? 'true' : undefined}
		class="flex h-11 w-full items-center justify-between gap-3 rounded-md border border-input bg-card px-3 text-left tabular-nums hover:bg-muted sm:w-80"
	>
		<span>{formatLocalInput(value)}</span>
		<CalendarDays class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	</Popover.Trigger>
	<!-- 그림자 없이 선으로 (DESIGN 6장 "모양 D2"). 생성 부품은 고치지 않고 class로 덮는다 -->
	<Popover.Content align="start" class="w-auto gap-3 border p-3 shadow-none ring-0">
		<!-- 생성 부품의 제목은 "월 연" 순서로 고정이라 "11월 2026년"이 된다. 고치지 않고 월 자리에 연 · 월을 다 넣고 연 자리를 비운다 -->
		<Calendar
			type="single"
			locale="ko-KR"
			calendarLabel="날짜"
			preventDeselect
			bind:placeholder={shown}
			monthFormat={(m) => `${shown.year}년 ${m}월`}
			yearFormat={() => ''}
			maxValue={today}
			value={parseDate(date)}
			onValueChange={(d) => {
				if (d) value = `${d.toString()}T${hour}:${minute}`;
			}}
			class="p-0"
		/>
		<div class="flex items-center gap-2 border-t pt-3">
			<select
				aria-label="시"
				class="h-11 flex-1 rounded-md border border-input bg-card px-2 tabular-nums"
				value={hour}
				onchange={(ev) => (value = `${date}T${ev.currentTarget.value}:${minute}`)}
			>
				{#each HOURS as h (h)}<option value={h}>{h}시</option>{/each}
			</select>
			<select
				aria-label="분"
				class="h-11 flex-1 rounded-md border border-input bg-card px-2 tabular-nums"
				value={minute}
				onchange={(ev) => (value = `${date}T${hour}:${ev.currentTarget.value}`)}
			>
				{#each MINUTES as m (m)}<option value={m}>{m}분</option>{/each}
			</select>
			<Button class="min-h-11" onclick={() => (open = false)}>확인</Button>
		</div>
	</Popover.Content>
</Popover.Root>
