package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mollshf/ums/internal/shared/utility"
)

type Response struct {
	Success bool              `json:"success" example:"true"`
	Data    any               `json:"data,omitempty"`
	Error   *utility.APIError `json:"error,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
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

func Failed(c *gin.Context, err *utility.APIError) {
	c.JSON(err.Status, &Response{
		Success: false,
		Error:   err,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, &Response{
		Success: true,
		Data:    data,
	})
}
