package user

import (
	"github.com/gin-gonic/gin"
	"github.com/mollshf/starter-kit/internal/shared/queries"
	"github.com/mollshf/starter-kit/internal/shared/utility"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRole(c *gin.Context) error {
	var createRoleRequest CreateRoleRequest
	err := c.ShouldBind(&createRoleRequest)
	if err != nil {
		return utility.ParseValidationError(err)
	}

	_, err = h.service.repo.InsertRole(c.Request.Context(), &RolePayload{
		RoleName:     createRoleRequest.RoleName,
		RoleCode:     createRoleRequest.RoleCode,
		RoleCategory: queries.RoleTypeSystem,
	})
	if err != nil {
		return err
	}
	return nil
}
