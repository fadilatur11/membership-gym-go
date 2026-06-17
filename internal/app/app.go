package app

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"membership-gym/config"
	"membership-gym/docs"
	"membership-gym/internal/middleware"
	"membership-gym/internal/modules"
	"membership-gym/internal/routes"
)

func New(cfg config.Config, db *pgxpool.Pool, redisClient *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware(), middleware.CORSMiddleware(), middleware.RecoveryMiddleware())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": true, "message": "OK", "data": gin.H{"status": "ok"}})
	})
	docs.SwaggerInfo.Host = "localhost:" + cfg.AppPort
	docs.SwaggerInfo.Schemes = []string{"http"}
	router.GET("/swagger-json/doc.json", func(ctx *gin.Context) {
		ctx.Data(200, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger-json/doc.json")))

	routes.Register(router, routes.Dependencies{
		Controllers:    modules.NewControllers(cfg, db, redisClient),
		AuthMiddleware: middleware.NewAuthMiddleware(cfg.JWTAccessSecret),
		RoleMiddleware: middleware.NewRoleMiddleware(),
	})
	return router
}
