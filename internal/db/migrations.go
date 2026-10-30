// Package db는 Postgres 연결, 마이그레이션 파일, sqlc 생성 코드(dbgen)를 담는다 (ARCHITECTURE 2절).
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	// 밑줄 import: 이름은 안 쓰고 "pgx"라는 database/sql 드라이버를 등록만 한다.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrations는 migrations 폴더의 SQL 파일들을 실행 파일 안에 넣은 것이다.
// 서버에 SQL 파일을 따로 올리지 않아도 app migrate가 이걸 읽는다 (ARCHITECTURE 1절).
// 비유: 도시락 통에 반찬까지 같이 싸 간다.
//
//go:embed migrations/*.sql
var Migrations embed.FS

// OpenMigrator는 마이그레이션을 실행할 goose Provider를 만든다. app migrate와 통합 테스트가 같이 쓴다.
// 돌려준 *sql.DB는 다 쓴 뒤 닫는다.
func OpenMigrator(databaseURL string) (*goose.Provider, *sql.DB, error) {
	// goose는 표준 database/sql 연결을 받는다. pgx가 그 어댑터(stdlib)를 같이 준다.
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres 연결: %w", err)
	}

	// embed한 파일들은 "migrations/00001_init.sql"처럼 폴더 이름이 붙어 있다.
	// goose는 맨 위에서 파일을 찾으므로 migrations 폴더 안을 맨 위로 잘라 넘긴다.
	files, err := fs.Sub(Migrations, "migrations")
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, files)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("goose 준비: %w", err)
	}
	return provider, sqlDB, nil
}
