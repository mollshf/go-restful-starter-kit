package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mollshf/starter-kit/internal/database"
	"github.com/mollshf/starter-kit/internal/shared/queries"
	"github.com/mollshf/starter-kit/internal/shared/web"
)

// @title			UMS API
// @version		1.0
// @description	User Management System API — Go modulith starter kit.
// @host			localhost:8080
// @BasePath		/api
// @schemes		http https
func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(corsConfig))

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	repo := queries.New(db)

	database := database.NewDatabase(db)
	router.GET("/health/db", web.Wrap(database.Health))

	ApiRoutes(router, db, repo)

	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
