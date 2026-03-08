package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "gorm.io/gorm"

	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/pkg/config"
	"backend/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	uploadDir := "uploads"
	os.MkdirAll(uploadDir, 0755)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	r.MaxMultipartMemory = 500 << 20 // 500 MB
	r.Static("/uploads", "./uploads")

	//authHandler := handlers.NewAuthHandler(db)
	newsHandler := handlers.NewNewsHandler(db)
	releaseHandler := handlers.NewReleaseHandler(db, uploadDir, baseURL)
	gameHandler := handlers.NewGameHandler(db)

	api := r.Group("/api")
	api.Use(middleware.RateLimitMiddleware())
	{
		v1 := api.Group("/v1")
		{
			newsPublic := v1.Group("/news")
			{
				newsPublic.GET("/", newsHandler.GetAllNews)
// 				newsPublic.GET("/:id", newsHandler.GetNewsByID)
// 				newsPublic.GET("/theme/:theme", newsHandler.GetNewsByTheme)
// 				newsPublic.GET("/search", newsHandler.SearchNews)
//
// 				newsPublic.POST("/", newsHandler.CreateNews)
// 				newsPublic.PUT("/:id", newsHandler.UpdateNews)
// 				newsPublic.DELETE("/:id", newsHandler.DeleteNews)
			}

			releasesPublic := v1.Group("/releases")
			{
				releasesPublic.GET("/", releaseHandler.GetAllReleases)
				releasesPublic.GET("/latest", releaseHandler.GetLatestRelease)
				releasesPublic.GET("/:id", releaseHandler.GetReleaseByID)
				releasesPublic.GET("/:id/download", releaseHandler.DownloadRelease)

				releasesPublic.POST("/", releaseHandler.CreateRelease)
				releasesPublic.PUT("/:id", releaseHandler.UpdateRelease)
				releasesPublic.DELETE("/:id", releaseHandler.DeleteRelease)
			}

			gamesPublic := v1.Group("/games")
			{
				gamesPublic.GET("/active", gameHandler.GetActiveGames)
// 				gamesPublic.GET("/stats", gameHandler.GetGamesStats)
//
// 				gamesPublic.POST("/", gameHandler.CreateGame)
// 				gamesPublic.POST("/:id/finish", gameHandler.FinishGame)
// 				gamesPublic.POST("/:id/cancel", gameHandler.CancelGame)
// 				gamesPublic.PATCH("/:id/players", gameHandler.UpdatePlayerCount)
			}

			// Health check
			v1.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status":   "ok",
					"database": "connected",
				})
			})
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
