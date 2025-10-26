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
	Count   int    `json:"count"`
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
		result, err := db.Exec("INSERT INTO tags (tag_name, count) VALUES (?, 0)", tagName)
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

	// 获取旧标签并减少计数
	rows, err := tx.Query(`
		SELECT tag_id FROM picture_tags WHERE picture_id = ?
	`, pictureID)
	if err != nil {
		return err
	}

	var oldTagIDs []int
	for rows.Next() {
		var tagID int
		if err := rows.Scan(&tagID); err != nil {
			rows.Close()
			return err
		}
		oldTagIDs = append(oldTagIDs, tagID)
	}
	rows.Close()

	// 删除旧关联
	if _, err := tx.Exec("DELETE FROM picture_tags WHERE picture_id = ?", pictureID); err != nil {
		return err
	}

	// 减少旧标签计数
	for _, tagID := range oldTagIDs {
		if _, err := tx.Exec("UPDATE tags SET count = count - 1 WHERE id = ?", tagID); err != nil {
			return err
		}
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

		// 增加标签计数
		if _, err := tx.Exec("UPDATE tags SET count = count + 1 WHERE id = ?", tagID); err != nil {
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

			// 增加标签计数
			if _, err := tx.Exec("UPDATE tags SET count = count + 1 WHERE id = ?", tagID); err != nil {
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
		result, err := tx.Exec("INSERT INTO tags (tag_name, count) VALUES (?, 0)", tagName)
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
	// 获取总数
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM tags WHERE count > 0").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取标签列表
	offset := (page - 1) * limit
	rows, err := db.Query(
		"SELECT id, tag_name, count FROM tags WHERE count > 0 ORDER BY count DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
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
