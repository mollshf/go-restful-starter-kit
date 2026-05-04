package web

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/mollshf/starter-kit/internal/shared/utility"
)

type AppHandler func(c *gin.Context) error

func Wrap(h AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {

		err := h(c)
		if err != nil {
			var apiError *utility.APIError
			if errors.As(err, &apiError) {
				slog.Warn("Kesalahan User", "error", err)
				Failed(c, apiError)
				return
			}

			slog.Error("Terjadi kesalahan pada internal server", "error", err)
			Failed(c, utility.NewInternalServerError("Terjadi kesalahan pada internal server", "INTERNAL_SERVER_ERROR"))
		}
	}
}
