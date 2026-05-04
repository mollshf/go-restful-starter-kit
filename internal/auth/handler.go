package auth

import (
	"fmt"
	"net/netip"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mollshf/ums/internal/shared/utility"
	"github.com/mollshf/ums/internal/shared/web"
)

type Person struct {
	ID   string `uri:"id" binding:"required,uuid"`
	Name string `uri:"name" binding:"required"`
}

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// CreateUser creates a user.
//
//	@Summary		Buat user
//	@Description	Menyimpan user baru
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateUserRequest	true	"Payload"
//	@Success		201		{object}	web.Response
//	@Failure		400		{object}	web.ErrorResponse
//	@Failure		500		{object}	web.ErrorResponse
//	@Router			/auth/register [post]
func (h *AuthHandler) RegisterUser(c *gin.Context) error {
	var createUserRequest CreateUserRequest
	err := c.ShouldBind(&createUserRequest)
	if err != nil {
		return utility.ParseValidationError(err)
	}

	_, err = h.authService.RegisterUser(c.Request.Context(), &createUserRequest)
	if err != nil {
		return err
	}

	web.Created(c, gin.H{
		"message": "User berhasil didaftarkan",
	})
	return nil

}

// Me returns the currently authenticated user's profile.
//
//	@Summary		Profil user aktif
//	@Description	Mengembalikan informasi user yang sedang login berdasarkan cookie sesi.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	web.Response
//	@Failure		401	{object}	web.ErrorResponse
//	@Failure		404	{object}	web.ErrorResponse
//	@Failure		500	{object}	web.ErrorResponse
//	@Router			/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) error {
	userIDValue, exists := c.Get(ContextUserIDKey)
	if !exists {
		return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return utility.NewInternalServerError("Format user_id pada konteks tidak valid", "AUTH_CONTEXT_INVALID")
	}

	user, err := h.authService.GetMe(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	web.OK(c, user)
	return nil
}

// ListActiveSessions returns all sessions currently held by the user.
//
//	@Summary		Daftar sesi aktif user
//	@Description	Mengembalikan seluruh sesi aktif (belum expired) milik user yang sedang login. Field is_current menandai sesi yang sedang membuat request ini.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	web.Response{data=[]SessionResponse}
//	@Failure		401	{object}	web.ErrorResponse
//	@Failure		500	{object}	web.ErrorResponse
//	@Router			/auth/sessions [get]
func (h *AuthHandler) ListActiveSessions(c *gin.Context) error {
	userIDValue, exists := c.Get(ContextUserIDKey)
	if !exists {
		return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return utility.NewInternalServerError("Format user_id pada konteks tidak valid", "AUTH_CONTEXT_INVALID")
	}

	currentSessionID, _ := c.Get(ContextSessionIDKey)
	currentID, _ := currentSessionID.(uuid.UUID)

	sessions, err := h.authService.ListActiveSessions(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	resp := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, SessionResponse{
			ID:         s.ID,
			IPAddress:  s.IpAddress,
			UserAgent:  s.UserAgent,
			ActiveRole: s.ActiveRole,
			CreatedAt:  s.CreatedAt,
			ExpiresAt:  s.ExpiresAt,
			IsCurrent:  s.ID == currentID,
		})
	}

	web.OK(c, resp)
	return nil
}

// LogoutAll terminates every active session for the current user.
//
//	@Summary		Logout dari semua perangkat
//	@Description	Menghapus seluruh sesi user yang sedang login dari semua perangkat. Berguna saat user mengganti password atau curiga akunnya disalahgunakan.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	web.Response
//	@Header			200	{string}	Set-Cookie	"Cookie sesi dihapus (X-academic-sesi=; Max-Age=-1)"
//	@Failure		401	{object}	web.ErrorResponse
//	@Failure		500	{object}	web.ErrorResponse
//	@Router			/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) error {
	userIDValue, exists := c.Get(ContextUserIDKey)
	if !exists {
		return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return utility.NewInternalServerError("Format user_id pada konteks tidak valid", "AUTH_CONTEXT_INVALID")
	}

	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		return err
	}

	clearCookie := web.GetClearCookieConfig()
	c.SetCookieData(&clearCookie)

	web.OK(c, gin.H{
		"message": "Logout dari semua perangkat berhasil",
	})
	return nil
}

// Logout clears the user session.
//
//	@Summary		Logout user
//	@Description	Menghapus sesi user dari database dan menghapus cookie sesi pada response.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	web.Response
//	@Header			200	{string}	Set-Cookie	"Cookie sesi dihapus (X-academic-sesi=; Max-Age=-1)"
//	@Failure		401	{object}	web.ErrorResponse
//	@Failure		500	{object}	web.ErrorResponse
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) error {
	rawToken, err := c.Cookie(web.AuthCookieName)
	if err != nil {
		return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
	}

	if err := h.authService.Logout(c.Request.Context(), rawToken); err != nil {
		return err
	}

	clearCookie := web.GetClearCookieConfig()
	c.SetCookieData(&clearCookie)

	web.OK(c, gin.H{
		"message": "Logout berhasil",
	})
	return nil
}

// Login authenticates a user.
//
//	@Summary		Login user
//	@Description	Autentikasi user dengan username dan password. Jika berhasil, cookie sesi akan di-set pada response.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"Kredensial login"
//	@Success		200		{object}	web.Response{data=MessageOnlyResponse}
//	@Header			200		{string}	Set-Cookie	"Cookie sesi (X-academic-sesi); HttpOnly; Path=/; Max-Age=86400"
//	@Failure		400		{object}	web.ErrorResponse
//	@Failure		401		{object}	web.ErrorResponse
//	@Failure		500		{object}	web.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) error {
	var loginRequest LoginRequest
	err := c.ShouldBind(&loginRequest)
	if err != nil {
		return utility.ParseValidationError(err)
	}

	// Convert string IP ke netip.Addr
	ipAddr, err := netip.ParseAddr(c.ClientIP())
	if err != nil {
		return fmt.Errorf("invalid IP address: %w", err)
	}

	loginResult, err := h.authService.Login(c.Request.Context(), &LoginInput{
		Username:  loginRequest.Username,
		Password:  loginRequest.Password,
		UserAgent: c.GetHeader("User-Agent"),
		IpAddress: ipAddr,
	})

	if err != nil {
		return err
	}

	// set cookie
	cookie := web.GetCookieConfig(loginResult.Token, loginResult.ExpiresAt)
	c.SetCookieData(&cookie)

	web.OK(c, &MessageOnlyResponse{
		Message: "Login berhasil",
	})
	return nil
}
