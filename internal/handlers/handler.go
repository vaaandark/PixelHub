package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to open file",
		})
		return
	}
	defer src.Close()

	// 计算文件哈希
	hasher := sha256.New()
	fileContent, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to read file",
		})
		return
	}
	hasher.Write(fileContent)
	hash := hex.EncodeToString(hasher.Sum(nil))

	// 生成唯一 ID（使用哈希的前 12 位）
	imageID := "img_" + hash[:12]

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
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: fmt.Sprintf("Failed to upload: %v", err),
		})
		return
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
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "Failed to save to database",
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "Upload successful",
		Data: map[string]interface{}{
			"image_id":    imageID,
			"url":         url,
			"hash":        hash,
			"description": description,
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

// MCPListTags MCP 服务：列出标签
func (h *Handler) MCPListTags(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	tags, total, err := database.ListTags(h.db, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tags",
		})
		return
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.TagName
	}

	hasMore := page*limit < total

	c.JSON(http.StatusOK, gin.H{
		"tags":     tagNames,
		"total":    total,
		"has_more": hasMore,
	})
}

// MCPSearchRelevance MCP 服务：相关性搜索
func (h *Handler) MCPSearchRelevance(c *gin.Context) {
	tagsParam := c.Query("tags")
	if tagsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tags parameter is required",
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}

	// 转换为 MCP 响应格式
	mcpResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		mcpResults[i] = map[string]interface{}{
			"image_id":          result.ID,
			"url":               result.URL,
			"description":       result.Description,
			"matched_tag_count": result.MatchedTagCount,
			"tags":              result.Tags,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"results": mcpResults,
	})
}

// AuthMiddleware MCP 鉴权中间件
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 支持 Bearer token 格式
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
