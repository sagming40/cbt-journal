import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// 프로젝트 코드는 Svelte 5 runes 모드로 강제한다(라이브러리는 제외). Svelte 6에서는 지워도 된다.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// SPA: 빌드하면 web/build/에 정적 파일만 나온다. 없는 경로는 index.html로 보낸다 (decisions 0003).
			// Go도 /api가 아닌 경로에 index.html을 준다 — 같은 규칙을 양쪽에 둔다 (ARCHITECTURE 1절).
			adapter: adapter({ fallback: 'index.html' })
		})
	],
	server: {
		// 포트를 고정한다: 5176이 차 있으면 다른 포트로 몰래 옮기지 말고 실패한다 (ARCHITECTURE 6절).
		port: 5176,
		strictPort: true,
		// 개발 중 /api 요청은 Go(8080)로 넘긴다. 브라우저에는 같은 출처로 보여서 CORS 설정이 필요 없다.
		proxy: {
			'/api': 'http://127.0.0.1:8080'
		}
	}
});
