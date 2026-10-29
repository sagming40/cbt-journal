// docs/mockups/shared.js
// 목업 공용: 컬러 시안 값, 테마 적용, 대비 계산, 상단 도구 막대.
// 주소 뒤에 ?palette=B&dark=1 처럼 붙이면 그 시안 · 테마로 그린다.

// ─── 무채색 (모든 시안 공통) ───────────────────────────────
const NEUTRAL = {
  light: { bg: '#F7F7F5', surface: '#FFFFFF', surface2: '#EFEFEC', border: '#E3E3DF', text: '#1B1B1F', muted: '#5F5F66' },
  dark:  { bg: '#111214', surface: '#191A1D', surface2: '#24252A', border: '#2E2F35', text: '#EDEDF0', muted: '#A3A3AD' },
};

// ─── 포인트 색 시안 (DESIGN 6장: 빨강 · 주황 · 초록 제외) ────────
// accent = 버튼 바탕, text = 배경 위 포인트 글자 (4.5:1 이상이어야 함)
const PALETTES = {
  A: { name: 'A · 깊은 청록', accent: '#2C7A7B', onAccent: '#FFFFFF', text: { light: '#22676A', dark: '#6CC5C4' } },
  B: { name: 'B · 차분한 남색', accent: '#4457C2', onAccent: '#FFFFFF', text: { light: '#3949A8', dark: '#9AA6F0' } },
  C: { name: 'C · 라벤더',     accent: '#7357B8', onAccent: '#FFFFFF', text: { light: '#64489F', dark: '#B9A5EE' } },
  D: { name: 'D · 모래',       accent: '#8A6D45', onAccent: '#FFFFFF', text: { light: '#75592F', dark: '#D2B48A' } },
};

const params = new URLSearchParams(location.search);
const state = {
  palette: PALETTES[params.get('palette')] ? params.get('palette') : 'A',
  dark: params.has('dark') ? params.get('dark') === '1' : matchMedia('(prefers-color-scheme: dark)').matches,
};

// 테마 변수를 요소에 심는다. 컬러 보드는 한 화면에 여러 시안을 그려서 요소 단위로 적용.
function applyTheme(el, paletteKey, dark) {
  const p = PALETTES[paletteKey];
  const n = dark ? NEUTRAL.dark : NEUTRAL.light;
  const vars = {
    '--bg': n.bg, '--surface': n.surface, '--surface-2': n.surface2, '--border': n.border,
    '--text': n.text, '--muted': n.muted,
    '--accent': p.accent, '--on-accent': p.onAccent, '--accent-text': dark ? p.text.dark : p.text.light,
    '--accent-tint': `color-mix(in oklab, ${p.accent} ${dark ? 18 : 9}%, ${n.surface})`,
    '--accent-line': `color-mix(in oklab, ${p.accent} ${dark ? 60 : 45}%, ${n.surface})`,
  };
  for (const [k, v] of Object.entries(vars)) el.style.setProperty(k, v);
  el.dataset.theme = dark ? 'dark' : 'light';
}

// 강도(0~100) → 포인트 색 명도 단계 (DESIGN 6장). 0도 바탕과 구분되게 15%에서 시작.
function intensityColor(v) {
  return `color-mix(in oklab, var(--accent) ${Math.round(15 + v * 0.85)}%, var(--surface-2))`;
}

// ─── WCAG 대비 ──────────────────────────────────────────────
function luminance(hex) {
  const c = hex.replace('#', '').match(/../g).map((h) => parseInt(h, 16) / 255)
    .map((x) => (x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4));
  return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2];
}
function contrast(a, b) {
  const [x, y] = [luminance(a), luminance(b)].sort((m, n) => n - m);
  return (x + 0.05) / (y + 0.05);
}

// ─── 상단 도구 막대 (보드 바깥) ────────────────────────────────
// extra: [{ label, options: [[값, 이름]...], value, onChange }]
function renderToolbar(onChange, extra = []) {
  const bar = document.createElement('div');
  bar.className = 'toolbar';
  const sel = document.createElement('select');
  for (const [k, p] of Object.entries(PALETTES)) sel.add(new Option(p.name, k, false, k === state.palette));
  const dark = document.createElement('label');
  dark.innerHTML = `<input type="checkbox" ${state.dark ? 'checked' : ''}> 다크`;
  bar.append('시안 ', sel, dark);
  for (const e of extra) {
    const s = document.createElement('select');
    for (const [v, name] of e.options) s.add(new Option(name, v, false, v === e.value));
    s.onchange = () => e.onChange(s.value);
    bar.append(` ${e.label} `, s);
  }
  const sync = () => {
    params.set('palette', state.palette);
    params.set('dark', state.dark ? '1' : '0');
    history.replaceState(null, '', '?' + params);
    onChange();
  };
  sel.onchange = () => { state.palette = sel.value; sync(); };
  dark.querySelector('input').onchange = (ev) => { state.dark = ev.target.checked; sync(); };
  document.body.prepend(bar);
}
