package shared

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool  `json:"success"`
	Data    any   `json:"data,omitempty"`
	Error   any   `json:"error,omitempty"`
	Meta    *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &Response{
		Success: true,
		Data:    data,
	})
}

func Failed(c *gin.Context, error *APIError) {
	c.JSON(error.Status, &Response{
		Success: false,
		Error:   error,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, &Response{
		Success: true,
		Data:    data,
	})
}
