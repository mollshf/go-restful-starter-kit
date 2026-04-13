package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authRepository *AuthRepository
}

func NewAuthHandler(authRepository *AuthRepository) *AuthHandler {
	return &AuthHandler{authRepository: authRepository}
}

// CreateUser creates a user.
//
//	@Summary		Buat user
//	@Description	Menyimpan user baru
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateUserRequest	true	"Payload"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	shared.APIError
//	@Router			/users [post]
func (h *AuthHandler) CreateUser(c *gin.Context) error {
	var createUserRequest CreateUserRequest
	err := c.ShouldBind(&createUserRequest)
	if err != nil {
		fmt.Println(err, "INI ADALAH ERRORNYA CUY")
		return err
	}

	id, err := h.authRepository.InsertUser(c.Request.Context(), &User{
		Name: createUserRequest.Name,
	})

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil diinput",
		"data":    id,
	})
	return nil

}

// GetUsers gets all users.
//
//	@Summary		Ambil semua user
//	@Description	Mengambil semua user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response
//	@Failure		400	{object}	shared.APIError
//	@Router			/users [get]
func (h *AuthHandler) GetUsers(c *gin.Context) error {

	users, err := h.authRepository.GetUsers(c.Request.Context())

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil diambil",
		"data":    users,
	})
	return nil

}

// DeleteUser deletes a user.
//
//	@Summary		Hapus user
//	@Description	Menghapus user berdasarkan ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	Response
//	@Failure		400	{object}	shared.APIError
//	@Router			/users/:id [delete]
func (h *AuthHandler) DeleteUser(c *gin.Context) error {
	id := c.Param("id")

	err := h.authRepository.DeleteUser(c.Request.Context(), uuid.MustParse(id))

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Berhasil dihapus",
	})
	return nil
}
