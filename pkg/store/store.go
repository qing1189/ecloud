package store

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"

	"ecloud_computer_auto_boot/pkg/util"
	_ "modernc.org/sqlite"
)

var db *sql.DB

const (
	DataDir    = "data"
	DBFileName = "ecloud.db"
)

// Init 初始化 SQLite 数据库（自动建表和迁移）
func Init(basePath string) error {
	if basePath == "" {
		basePath = "."
	}
	dataDir := filepath.Join(basePath, DataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, DBFileName)
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return err
	}
	db = conn

	// 设置连接池（SQLite 不适合多写并发，控制住）
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := migrate(); err != nil {
		return err
	}

	util.Log().Info("[存储] SQLite 数据库已初始化: %s", dbPath)
	return nil
}

// DB 返回数据库连接
func DB() *sql.DB {
	return db
}

// migrate 自动建表
func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id            TEXT PRIMARY KEY,
		username      TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role          TEXT NOT NULL DEFAULT 'user',
		display_name  TEXT DEFAULT '',
		email         TEXT DEFAULT '',
		created_at    TEXT DEFAULT (datetime('now','localtime')),
		updated_at    TEXT DEFAULT (datetime('now','localtime')),
		last_login_at TEXT DEFAULT ''
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username COLLATE NOCASE);

	CREATE TABLE IF NOT EXISTS accounts (
		id            TEXT PRIMARY KEY,
		name          TEXT NOT NULL,
		type          TEXT NOT NULL DEFAULT 'public',
		username      TEXT DEFAULT '',
		password      TEXT DEFAULT '',
		access_key    TEXT DEFAULT '',
		secret_key    TEXT DEFAULT '',
		pool_id       TEXT DEFAULT '',
		monitor_enabled  INTEGER DEFAULT 0,
		monitor_interval INTEGER DEFAULT 60,
		monitor_machines TEXT DEFAULT '[]',
		user_id       TEXT DEFAULT '',
		created_at    TEXT DEFAULT (datetime('now','localtime')),
		updated_at    TEXT DEFAULT (datetime('now','localtime'))
	);

	CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);

	CREATE TABLE IF NOT EXISTS auth (
		id            INTEGER PRIMARY KEY CHECK (id = 1),
		password_hash TEXT NOT NULL DEFAULT '',
		secret_key    TEXT NOT NULL DEFAULT '',
		created_at    TEXT DEFAULT (datetime('now','localtime'))
	);

	CREATE TABLE IF NOT EXISTS logs (
		id            TEXT PRIMARY KEY,
		timestamp     TEXT DEFAULT (datetime('now','localtime')),
		type          TEXT NOT NULL,
		account_id    TEXT DEFAULT '',
		user_id       TEXT DEFAULT '',
		message       TEXT NOT NULL DEFAULT '',
		details       TEXT DEFAULT '',
		status        TEXT NOT NULL DEFAULT 'info'
	);

	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_type     ON logs(type);
	CREATE INDEX IF NOT EXISTS idx_logs_account  ON logs(account_id);
	CREATE INDEX IF NOT EXISTS idx_logs_user_id  ON logs(user_id);
	`

	_, err := db.Exec(schema)
	return err
}

// JSONString 将对象转为 JSON 字符串
func JSONString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}