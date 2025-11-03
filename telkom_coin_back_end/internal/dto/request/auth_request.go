package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Pin      string `json:"pin"`
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePinRequest struct {
	OldPin string `json:"old_pin" binding:"required,len=6,numeric"`
	NewPin string `json:"new_pin" binding:"required,len=6,numeric"`
}

func (r RegisterRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username,
			validation.Required.Error("Username wajib diisi"),
			validation.Length(3, 50).Error("Username harus 3–50 karakter"),
		),
		validation.Field(&r.Email,
			validation.Required.Error("Email wajib diisi"),
			is.Email.Error("Format email tidak valid"),
		),
		validation.Field(&r.Phone,
			validation.Required.Error("Nomor telepon wajib diisi"),
			validation.Length(10, 20).Error("Nomor telepon harus 10–20 digit"),
		),
		validation.Field(&r.Password,
			validation.Required.Error("Password wajib diisi"),
			validation.Length(6, 0).Error("Password minimal 6 karakter"),
		),
		validation.Field(&r.Pin,
			validation.Required.Error("Pin wajib diisi"),
		),
	)
}