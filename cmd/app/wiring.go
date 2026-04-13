package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/academic/internal/auth"
)

type Wiring struct {
	*auth.AuthHandler
}

func NewWiring(db *pgxpool.Pool) (*Wiring, error) {
	authRepository := auth.NewAuthRepository(db)
	authHandler := auth.NewAuthHandler(authRepository)
	return &Wiring{
		AuthHandler: authHandler,
	}, nil
}
