package database

import (
	"database/sql"
	"time"
)

type Picture struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	StorageKey  string    `json:"storage_key"`
	Hash        string    `json:"hash"`
	Description string    `json:"description"`
	UploadDate  time.Time `json:"upload_date"`
	Deleted     bool      `json:"deleted"`
}

type Tag struct {
	ID      int    `json:"id"`
	TagName string `json:"tag_name"`
	Count   int    `json:"count"` // 动态计算，不存储在数据库
}

type PictureWithTags struct {
	Picture
	Tags            []string `json:"tags"`
	MatchedTagCount int      `json:"matched_tag_count,omitempty"`
}

// CreatePicture 创建新图片记录
func CreatePicture(db *sql.DB, pic *Picture) error {
	_, err := db.Exec(
		"INSERT INTO pictures (id, url, storage_key, hash, description) VALUES (?, ?, ?, ?, ?)",
		pic.ID, pic.URL, pic.StorageKey, pic.Hash, pic.Description,
	)
	return err
}

// GetPicture 获取图片详情
func GetPicture(db *sql.DB, id string) (*Picture, error) {
	var pic Picture
	err := db.QueryRow(
		"SELECT id, url, storage_key, hash, description, upload_date, deleted FROM pictures WHERE id = ? AND deleted = 0",
		id,
	).Scan(&pic.ID, &pic.URL, &pic.StorageKey, &pic.Hash, &pic.Description, &pic.UploadDate, &pic.Deleted)

	if err != nil {
		return nil, err
	}
	return &pic, nil
}

// ListPictures 列出所有图片（分页）
func ListPictures(db *sql.DB, page, limit int, sort string) ([]PictureWithTags, int, error) {
	// 确定排序方式
	orderBy := "p.upload_date DESC" // 默认最新优先
	if sort == "date_asc" {
		orderBy = "p.upload_date ASC"
	}

	// 获取总数
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM pictures WHERE deleted = 0").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取图片列表
	offset := (page - 1) * limit
	query := `
		SELECT p.id, p.url, p.storage_key, p.hash, p.description, p.upload_date
		FROM pictures p
		WHERE p.deleted = 0
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []PictureWithTags
	for rows.Next() {
		var pic PictureWithTags
		if err := rows.Scan(&pic.ID, &pic.URL, &pic.StorageKey, &pic.Hash, &pic.Description, &pic.UploadDate); err != nil {
			return nil, 0, err
		}

		// 获取标签
		tags, err := GetPictureTags(db, pic.ID)
		if err != nil {
			return nil, 0, err
		}
		pic.Tags = tags

		results = append(results, pic)
	}

	return results, total, nil
}

// DeletePicture 软删除图片
func DeletePicture(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE pictures SET deleted = 1 WHERE id = ?", id)
	return err
}

// UpdatePictureDescription 更新图片描述
func UpdatePictureDescription(db *sql.DB, id string, description string) error {
	_, err := db.Exec("UPDATE pictures SET description = ? WHERE id = ? AND deleted = 0", description, id)
	return err
}

// GetOrCreateTag 获取或创建标签
func GetOrCreateTag(db *sql.DB, tagName string) (int, error) {
	var tagID int
	err := db.QueryRow("SELECT id FROM tags WHERE tag_name = ?", tagName).Scan(&tagID)

	if err == sql.ErrNoRows {
		// 标签不存在，创建新标签
		result, err := db.Exec("INSERT INTO tags (tag_name) VALUES (?)", tagName)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return int(id), err
	}

	return tagID, err
}

// SetPictureTags 设置图片的标签（替换所有）
func SetPictureTags(db *sql.DB, pictureID string, tagNames []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除旧关联
	if _, err := tx.Exec("DELETE FROM picture_tags WHERE picture_id = ?", pictureID); err != nil {
		return err
	}

	// 添加新标签
	for _, tagName := range tagNames {
		tagID, err := getOrCreateTagTx(tx, tagName)
		if err != nil {
			return err
		}

		// 创建新关联
		if _, err := tx.Exec("INSERT INTO picture_tags (picture_id, tag_id) VALUES (?, ?)", pictureID, tagID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AppendPictureTags 追加图片的标签
func AppendPictureTags(db *sql.DB, pictureID string, tagNames []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, tagName := range tagNames {
		tagID, err := getOrCreateTagTx(tx, tagName)
		if err != nil {
			return err
		}

		// 检查是否已存在
		var exists int
		err = tx.QueryRow("SELECT COUNT(*) FROM picture_tags WHERE picture_id = ? AND tag_id = ?", pictureID, tagID).Scan(&exists)
		if err != nil {
			return err
		}

		if exists == 0 {
			// 创建新关联
			if _, err := tx.Exec("INSERT INTO picture_tags (picture_id, tag_id) VALUES (?, ?)", pictureID, tagID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func getOrCreateTagTx(tx *sql.Tx, tagName string) (int, error) {
	var tagID int
	err := tx.QueryRow("SELECT id FROM tags WHERE tag_name = ?", tagName).Scan(&tagID)

	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO tags (tag_name) VALUES (?)", tagName)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return int(id), err
	}

	return tagID, err
}

// GetPictureTags 获取图片的所有标签
func GetPictureTags(db *sql.DB, pictureID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.tag_name 
		FROM tags t
		JOIN picture_tags pt ON t.id = pt.tag_id
		WHERE pt.picture_id = ?
	`, pictureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// ListTags 列出所有标签（分页）
func ListTags(db *sql.DB, page, limit int) ([]Tag, int, error) {
	// 获取总数（只统计有图片关联的标签）
	var total int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT t.id) 
		FROM tags t
		JOIN picture_tags pt ON t.id = pt.tag_id
		JOIN pictures p ON pt.picture_id = p.id
		WHERE p.deleted = 0
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取标签列表，按使用次数降序排列
	offset := (page - 1) * limit
	rows, err := db.Query(`
		SELECT t.id, t.tag_name, COUNT(pt.picture_id) as count
		FROM tags t
		JOIN picture_tags pt ON t.id = pt.tag_id
		JOIN pictures p ON pt.picture_id = p.id
		WHERE p.deleted = 0
		GROUP BY t.id, t.tag_name
		ORDER BY count DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.TagName, &tag.Count); err != nil {
			return nil, 0, err
		}
		tags = append(tags, tag)
	}

	return tags, total, nil
}

// SearchExact 精确搜索（AND 逻辑）
func SearchExact(db *sql.DB, tagNames []string, page, limit int) ([]PictureWithTags, int, error) {
	if len(tagNames) == 0 {
		return []PictureWithTags{}, 0, nil
	}

	// 构建查询
	query := `
		SELECT p.id, p.url, p.storage_key, p.hash, p.description, p.upload_date
		FROM pictures p
		WHERE p.deleted = 0 AND p.id IN (
			SELECT pt.picture_id
			FROM picture_tags pt
			JOIN tags t ON pt.tag_id = t.id
			WHERE t.tag_name IN (` + placeholders(len(tagNames)) + `)
			GROUP BY pt.picture_id
			HAVING COUNT(DISTINCT t.id) = ?
		)
		ORDER BY p.upload_date DESC
		LIMIT ? OFFSET ?
	`

	args := make([]interface{}, 0, len(tagNames)+3)
	for _, tag := range tagNames {
		args = append(args, tag)
	}
	args = append(args, len(tagNames))
	args = append(args, limit)
	args = append(args, (page-1)*limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []PictureWithTags
	for rows.Next() {
		var pic PictureWithTags
		if err := rows.Scan(&pic.ID, &pic.URL, &pic.StorageKey, &pic.Hash, &pic.Description, &pic.UploadDate); err != nil {
			return nil, 0, err
		}

		// 获取标签
		tags, err := GetPictureTags(db, pic.ID)
		if err != nil {
			return nil, 0, err
		}
		pic.Tags = tags

		results = append(results, pic)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM pictures p
		JOIN picture_tags pt ON p.id = pt.picture_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE p.deleted = 0 AND t.tag_name IN (` + placeholders(len(tagNames)) + `)
		GROUP BY p.id
		HAVING COUNT(DISTINCT t.id) = ?
	`
	countArgs := make([]interface{}, 0, len(tagNames)+1)
	for _, tag := range tagNames {
		countArgs = append(countArgs, tag)
	}
	countArgs = append(countArgs, len(tagNames))

	var total int
	err = db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return results, total, nil
}

// SearchRelevance 相关性搜索（OR 逻辑，按匹配数排序）
func SearchRelevance(db *sql.DB, tagNames []string, page, limit int) ([]PictureWithTags, int, error) {
	if len(tagNames) == 0 {
		return []PictureWithTags{}, 0, nil
	}

	query := `
		SELECT p.id, p.url, p.storage_key, p.hash, p.description, p.upload_date, COUNT(DISTINCT pt.tag_id) as matched_count
		FROM pictures p
		JOIN picture_tags pt ON p.id = pt.picture_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE p.deleted = 0 AND t.tag_name IN (` + placeholders(len(tagNames)) + `)
		GROUP BY p.id
		ORDER BY matched_count DESC, p.upload_date DESC
		LIMIT ? OFFSET ?
	`

	args := make([]interface{}, 0, len(tagNames)+2)
	for _, tag := range tagNames {
		args = append(args, tag)
	}
	args = append(args, limit)
	args = append(args, (page-1)*limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []PictureWithTags
	for rows.Next() {
		var pic PictureWithTags
		if err := rows.Scan(&pic.ID, &pic.URL, &pic.StorageKey, &pic.Hash, &pic.Description, &pic.UploadDate, &pic.MatchedTagCount); err != nil {
			return nil, 0, err
		}

		// 获取标签
		tags, err := GetPictureTags(db, pic.ID)
		if err != nil {
			return nil, 0, err
		}
		pic.Tags = tags

		results = append(results, pic)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM pictures p
		JOIN picture_tags pt ON p.id = pt.picture_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE p.deleted = 0 AND t.tag_name IN (` + placeholders(len(tagNames)) + `)
	`
	countArgs := make([]interface{}, 0, len(tagNames))
	for _, tag := range tagNames {
		countArgs = append(countArgs, tag)
	}

	var total int
	err = db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return results, total, nil
}

func placeholders(n int) string {
	if n == 0 {
		return ""
	}
	s := "?"
	for i := 1; i < n; i++ {
		s += ",?"
	}
	return s
}
