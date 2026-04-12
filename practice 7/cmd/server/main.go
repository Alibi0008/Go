package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"practice7/internal/controller/http/middleware"
	v1 "practice7/internal/controller/http/v1"
	"practice7/internal/usecase"
	"practice7/internal/usecase/repo"
	"practice7/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	repository := repo.NewUserRepo()
	jwtManager := utils.NewJWTManager(os.Getenv("JWT_SECRET"), 24*time.Hour)
	userUseCase := usecase.NewUserUseCase(repository, jwtManager)

	adminCreated, adminUser, err := userUseCase.EnsureBootstrapAdmin(
		envOrDefault("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		envOrDefault("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com"),
		envOrDefault("BOOTSTRAP_ADMIN_PASSWORD", "admin123"),
	)
	if err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}

	if adminCreated {
		log.Printf(
			"bootstrap admin created: username=%s password=%s id=%s",
			envOrDefault("BOOTSTRAP_ADMIN_USERNAME", "admin"),
			envOrDefault("BOOTSTRAP_ADMIN_PASSWORD", "admin123"),
			adminUser.ID,
		)
	} else {
		log.Printf("bootstrap admin already exists: id=%s", adminUser.ID)
	}

	router := gin.Default()
	router.Use(middleware.NewRateLimiter(
		jwtManager,
		envAsInt("RATE_LIMIT_MAX_REQUESTS", 5),
		time.Duration(envAsInt("RATE_LIMIT_WINDOW_SECONDS", 60))*time.Second,
	))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1.NewUserRoutes(router.Group(""), userUseCase, jwtManager)

	addr := ":" + envOrDefault("PORT", "8080")
	log.Printf("server is running on %s", addr)
	log.Fatal(router.Run(addr))
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func envAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
