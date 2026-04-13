package auth

type CreateUserRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	CsrfToken string   `json:"csrf_token"`
	Roles     []string `json:"roles"`
}
