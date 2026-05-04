package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mollshf/starter-kit/internal/shared/queries"
	"github.com/mollshf/starter-kit/internal/shared/utility"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRole(ctx context.Context, role *CreateRoleRequest) (*uuid.UUID, error) {
	roleID, err := s.repo.InsertRole(ctx, &RolePayload{
		RoleName:     role.RoleName,
		RoleCode:     role.RoleCode,
		RoleCategory: queries.RoleTypeSystem,
	})

	if err != nil {
		if errors.Is(err, ErrRoleNameDuplicate) {
			return nil, utility.NewConflictError("Nama role sudah ada", "ROLE_NAME_DUPLICATE")
		}
		if errors.Is(err, ErrRoleCodeDuplicate) {
			return nil, utility.NewConflictError("Kode role sudah ada", "ROLE_CODE_DUPLICATE")
		}
		return nil, utility.NewInternalServerError("Gagal membuat role", "FAILED_TO_CREATE_ROLE")
	}

	return roleID, nil
}
