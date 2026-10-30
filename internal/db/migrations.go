// Package db는 Postgres 연결, 마이그레이션 파일, (M2부터) sqlc 생성 코드를 담는다 (ARCHITECTURE 2절).
package db

import "embed"

// Migrations는 migrations 폴더의 SQL 파일들을 실행 파일 안에 넣은 것이다.
// 서버에 SQL 파일을 따로 올리지 않아도 app migrate가 이걸 읽는다 (ARCHITECTURE 1절).
// 비유: 도시락 통에 반찬까지 같이 싸 간다.
//
//go:embed migrations/*.sql
var Migrations embed.FS
