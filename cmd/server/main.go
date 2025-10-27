package main

import (
	"log"

	"github.com/vaaandark/PixelHub/internal/config"
	"github.com/vaaandark/PixelHub/internal/database"
	"github.com/vaaandark/PixelHub/internal/handlers"
	"github.com/vaaandark/PixelHub/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.InitDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 初始化存储提供商
	storageProvider, err := storage.NewProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage provider: %v", err)
	}

	// 创建 Gin 路由器
	r := gin.Default()

	// 设置 CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件服务（前端）
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")

	// 初始化处理器
	h := handlers.NewHandler(db, storageProvider)

	// 图床后端 API 路由
	api := r.Group("/api/v1")
	{
		// 图片管理
		api.POST("/images/upload", h.UploadImage)
		api.POST("/images/batch-upload", h.BatchUploadImages)
		api.GET("/images", h.ListImages)
		api.GET("/images/:image_id", h.GetImageDetail)
		api.PUT("/images/:image_id", h.UpdateImageDescription)
		api.DELETE("/images/:image_id", h.DeleteImage)

		// 标签管理
		api.PUT("/images/:image_id/tags", h.UpdateImageTags)
		api.GET("/tags", h.ListTags)

		// 搜索
		api.GET("/search/exact", h.SearchExact)
		api.GET("/search/relevance", h.SearchRelevance)
	}

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting PixelHub server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
