package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mollshf/academic/internal/auth"
	"github.com/mollshf/academic/internal/database"
	"github.com/mollshf/academic/internal/shared"
)

// @title           Academic API
// @version         1.0
// @description     API akademik.
// @host            localhost:8080
// @BasePath        /
func main() {
	godotenv.Load()
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(corsConfig))

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer db.Close()

	authRepository := auth.NewAuthRepository(db)
	authHandler := auth.NewAuthHandler(authRepository)
	database := database.NewDatabase(db)

	router.With()

	router.GET("/health/db", shared.Wrap(database.Health))

	router.POST("/users", shared.Wrap(authHandler.CreateUser))

	router.GET("/users", shared.Wrap(authHandler.GetUsers))

	router.DELETE("/users/:id", shared.Wrap(authHandler.DeleteUser))

	router.Run(":8080")
}
