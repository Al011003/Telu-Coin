package repository

import (
	"telkom_coin_back_end/internal/models"

	"gorm.io/gorm"
)

// UserRepositoryInterface defines the contract for user repository
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id int64) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByWalletAddress(address string) (*models.User, error)
	Update(user *models.User) error
	UpdateFields(id int64, fields map[string]interface{}) error
	Delete(id int64) error
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Get user by ID
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by wallet address
func (r *UserRepository) GetByWalletAddress(address string) (*models.User, error) {
	var user models.User
	err := r.db.Where("wallet_address = ?", address).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Update specific fields
func (r *UserRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete user (soft delete)
func (r *UserRepository) Delete(id int64) error {
	return r.db.Delete(&models.User{}, id).Error
}

// Check if email exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Check if username exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
