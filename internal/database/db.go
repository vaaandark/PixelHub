package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB 初始化数据库连接并创建表
func InitDB(dbPath string) (*sql.DB, error) {
	// 确保数据目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS pictures (
		id TEXT PRIMARY KEY,
		url TEXT NOT NULL,
		storage_key TEXT NOT NULL,
		hash TEXT NOT NULL,
		upload_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_name TEXT UNIQUE NOT NULL,
		count INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS picture_tags (
		picture_id TEXT NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (picture_id, tag_id),
		FOREIGN KEY (picture_id) REFERENCES pictures(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_picture_tags_picture ON picture_tags(picture_id);
	CREATE INDEX IF NOT EXISTS idx_picture_tags_tag ON picture_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(tag_name);
	`

	_, err := db.Exec(schema)
	return err
}
