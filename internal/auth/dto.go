package auth

type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	CsrToken string   `json:"csr_token"`
	Roles    []string `json:"roles"`
}
