package user

type CreateRoleRequest struct {
	RoleName string `json:"role_name" binding:"required,min=3,max=64"`
	RoleCode string `json:"role_code" binding:"required,min=1,max=32"`
}
