<script lang="ts">
	// 전 → 후 강도 (DESIGN 5.2절). 줄었다고 초록, 늘었다고 빨강을 쓰지 않는다 — 무채색 + 숫자만 (1장 원칙 2).
	let { before, after }: { before: number; after: number } = $props();
</script>

<span class="tabular-nums">
	<span class="sr-only">전 {before}, 후 {after}</span>
	<span aria-hidden="true">{before} → {after}</span>
</span>
