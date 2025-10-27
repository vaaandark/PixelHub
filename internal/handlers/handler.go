package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vaaandark/PixelHub/internal/database"
	"github.com/vaaandark/PixelHub/internal/storage"
)

type Handler struct {
	db      *sql.DB
	storage storage.Provider
}

func NewHandler(db *sql.DB, storageProvider storage.Provider) *Handler {
	return &Handler{
		db:      db,
		storage: storageProvider,
	}
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// uploadSingleImage 上传单个图片的核心逻辑（辅助函数）
func (h *Handler) uploadSingleImage(file *multipart.FileHeader, description string) (map[string]interface{}, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// 计算文件哈希
	hasher := sha256.New()
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	hasher.Write(fileContent)
	hash := hex.EncodeToString(hasher.Sum(nil))

	// 生成唯一 ID（使用 UUID）
	imageID := "img_" + uuid.New().String()

	// 生成存储 key
	ext := filepath.Ext(file.Filename)
	storageKey := imageID + ext

	// 上传到存储
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 重新创建 reader
	reader := strings.NewReader(string(fileContent))
	storageKey, url, err := h.storage.Upload(storageKey, reader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %v", err)
	}

	// 保存到数据库
	pic := &database.Picture{
		ID:          imageID,
		URL:         url,
		StorageKey:  storageKey,
		Hash:        hash,
		Description: description,
	}

	if err := database.CreatePicture(h.db, pic); err != nil {
		// 如果数据库保存失败，尝试删除已上传的文件
		h.storage.Delete(storageKey)
		return nil, fmt.Errorf("failed to save to database: %v", err)
	}

	return map[string]interface{}{
		"image_id":    imageID,
		"url":         url,
		"hash":        hash,
		"description": description,
	}, nil
}

// UploadImage 上传图片
func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "No file uploaded",
		})
		return
	}

	// 获取可选的描述信息
	description := c.PostForm("description")

	result, err := h.uploadSingleImage(file, description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "Upload successful",
		Data:    result,
	})
}

// BatchUploadImages 批量上传图片
func (h *Handler) BatchUploadImages(c *gin.Context) {
	// 获取 multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Failed to parse multipart form",
		})
		return
	}

	// 获取所有文件
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "No files uploaded",
		})
		return
	}

	// 批量上传结果
	type UploadResult struct {
		Filename string `json:"filename"`
		Status   string `json:"status"`
		ImageID  string `json:"image_id,omitempty"`
		URL      string `json:"url,omitempty"`
		Hash     string `json:"hash,omitempty"`
		Error    string `json:"error,omitempty"`
	}

	results := make([]UploadResult, 0, len(files))
	successCount := 0
	failedCount := 0

	// 逐个处理文件
	for _, file := range files {
		result := UploadResult{
			Filename: file.Filename,
		}

		// 上传文件（不带 description）
		data, err := h.uploadSingleImage(file, "")
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			failedCount++
		} else {
			result.Status = "success"
			result.ImageID = data["image_id"].(string)
			result.URL = data["url"].(string)
			result.Hash = data["hash"].(string)
			successCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Batch upload completed",
		Data: map[string]interface{}{
			"total":   len(files),
			"success": successCount,
			"failed":  failedCount,
			"results": results,
		},
	})
}

// DeleteImage 删除图片
func (h *Handler) DeleteImage(c *gin.Context) {
	imageID := c.Param("image_id")

	// 获取图片信息
	pic, err := database.GetPicture(h.db, imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: "Image not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to get image",
		})
		return
	}

	// 软删除数据库记录
	if err := database.DeletePicture(h.db, imageID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to delete image",
		})
		return
	}

	// 从存储删除文件（可选，根据需求决定是否立即删除）
	if err := h.storage.Delete(pic.StorageKey); err != nil {
		// 记录错误但不返回失败，因为数据库已经标记为删除
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Image deleted successfully",
	})
}

// ListImages 列出所有图片
func (h *Handler) ListImages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "date_desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 验证排序参数
	if sort != "date_desc" && sort != "date_asc" {
		sort = "date_desc"
	}

	images, total, err := database.ListPictures(h.db, page, limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to list images",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"total":        total,
			"current_page": page,
			"images":       images,
		},
	})
}

// GetImageDetail 获取图片详情
func (h *Handler) GetImageDetail(c *gin.Context) {
	imageID := c.Param("image_id")

	pic, err := database.GetPicture(h.db, imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: "Image not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to get image",
		})
		return
	}

	// 获取标签
	tags, err := database.GetPictureTags(h.db, imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to get tags",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"image_id":    pic.ID,
			"url":         pic.URL,
			"hash":        pic.Hash,
			"description": pic.Description,
			"upload_date": pic.UploadDate,
			"tags":        tags,
		},
	})
}

// UpdateImageTags 更新图片标签
func (h *Handler) UpdateImageTags(c *gin.Context) {
	imageID := c.Param("image_id")

	var req struct {
		Tags []string `json:"tags" binding:"required"`
		Mode string   `json:"mode"` // "set" or "append"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Invalid request",
		})
		return
	}

	// 默认为 set 模式
	if req.Mode == "" {
		req.Mode = "set"
	}

	// 检查图片是否存在
	_, err := database.GetPicture(h.db, imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: "Image not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to get image",
		})
		return
	}

	// 更新标签
	if req.Mode == "append" {
		err = database.AppendPictureTags(h.db, imageID, req.Tags)
	} else {
		err = database.SetPictureTags(h.db, imageID, req.Tags)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: fmt.Sprintf("Failed to update tags: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Tags updated successfully",
	})
}

// UpdateImageDescription 更新图片描述
func (h *Handler) UpdateImageDescription(c *gin.Context) {
	imageID := c.Param("image_id")

	var req struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Invalid request",
		})
		return
	}

	// 检查图片是否存在
	_, err := database.GetPicture(h.db, imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: "Image not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to get image",
		})
		return
	}

	// 更新描述
	err = database.UpdatePictureDescription(h.db, imageID, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: fmt.Sprintf("Failed to update description: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Description updated successfully",
	})
}

// ListTags 列出所有标签
func (h *Handler) ListTags(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	tags, total, err := database.ListTags(h.db, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to list tags",
		})
		return
	}

	// 转换为 API 响应格式
	tagList := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		tagList[i] = map[string]interface{}{
			"name":  tag.TagName,
			"count": tag.Count,
		}
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"total":        total,
			"current_page": page,
			"tags":         tagList,
		},
	})
}

// SearchExact 精确搜索（AND）
func (h *Handler) SearchExact(c *gin.Context) {
	tagsParam := c.Query("tags")
	if tagsParam == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Tags parameter is required",
		})
		return
	}

	tags := strings.Split(tagsParam, ",")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	results, total, err := database.SearchExact(h.db, tags, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Search failed",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"total":        total,
			"current_page": page,
			"results":      results,
		},
	})
}

// SearchRelevance 相关性搜索（OR）
func (h *Handler) SearchRelevance(c *gin.Context) {
	tagsParam := c.Query("tags")
	if tagsParam == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Tags parameter is required",
		})
		return
	}

	tags := strings.Split(tagsParam, ",")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	results, total, err := database.SearchRelevance(h.db, tags, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Search failed",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"total":        total,
			"current_page": page,
			"results":      results,
		},
	})
}
