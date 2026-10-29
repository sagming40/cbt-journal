# 0003. 프론트엔드는 SvelteKit SPA + shadcn-svelte로 만든다

- 상태: 채택
- 날짜: 2026-10-29

## 맥락
- 사용자는 React와 Vue를 써 봤고, 생태계 때문에 React만 반복해서 쓰는 데 질렸다.
- 조건은 하나: **shadcn 계열 UI를 쓸 수 있어야 한다.**
- 백엔드가 Go라서([0002](0002-backend-go.md)) 프론트를 위한 Node 서버를 운영에 따로 두고 싶지 않다.

## 결정
- **SvelteKit(Svelte 5)** 을 `adapter-static`으로 빌드해 SPA(정적 파일)로 만들고, Go 바이너리에 `go:embed`로 넣는다. 운영에는 실행 파일 하나만 있다.
- UI는 **shadcn-svelte + Tailwind CSS**, 아이콘은 Lucide.
- 언어 TypeScript, 패키지 매니저 `pnpm`, 폼은 `sveltekit-superforms` + Zod, 차트는 shadcn-svelte charts(LayerChart).
- API 호출은 OpenAPI 스펙에서 생성한 `openapi-typescript` + `openapi-fetch` 클라이언트로만 한다.

## 검토한 대안
| 후보 | 탈락 이유 |
|---|---|
| React / Vue | 사용자가 이미 써 봤고 질렸다 |
| Solid + solid-ui | JSX라 React와 너무 닮았고 포트 완성도가 낮다 |
| Angular + spartan/ui | 무겁고 shadcn 원본과 거리가 있다 |
| Go 템플릿 + htmx + Basecoat | JS가 거의 없어 가볍지만 슬라이더 · 차트처럼 상호작용 많은 화면에서 손이 더 간다 |
| Next.js 풀스택 | 백엔드를 Go로 나누기로 해서 맞지 않는다 |

## 감수할 것
- shadcn-svelte는 React 원본보다 업데이트가 늦다. 새 컴포넌트가 바로 안 들어올 수 있다.
- SSR이 없다. 로그인 뒤에만 쓰는 앱이라 검색 노출 · 첫 화면 속도 손해는 작다고 본다.
