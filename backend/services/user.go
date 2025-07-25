package services

import (
	"errors"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	dto "stream-demo/backend/dto"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"stream-demo/backend/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用戶服務
type UserService struct {
	Conf      *config.Config
	Repo      *postgresqlRepo.PostgreSQLRepo
	RepoSlave *postgresqlRepo.PostgreSQLRepo
	JWTUtil   *utils.JWTUtil
}

// NewUserService 創建用戶服務實例
func NewUserService(conf *config.Config) *UserService {
	return &UserService{
		Conf:      conf,
		Repo:      postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave: postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
		JWTUtil:   utils.NewJWTUtil(conf.JWT.Secret),
	}
}

// Register 用戶註冊
func (s *UserService) Register(username string, email string, password string) (*dto.UserDTO, error) {
	// 檢查用戶名是否已存在（讀操作 - 使用從庫）
	if _, err := s.RepoSlave.FindUserByUsername(username); err == nil {
		return nil, errors.New("用戶名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 檢查郵箱是否已存在（讀操作 - 使用從庫）
	if _, err := s.RepoSlave.FindUserByEmail(email); err == nil {
		return nil, errors.New("郵箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 創建用戶（寫操作 - 使用主庫）
	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.Repo.CreateUser(user); err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Login 用戶登入
func (s *UserService) Login(username string, password string) (string, *dto.UserDTO, time.Time, error) {
	// 查找用戶（讀操作 - 使用從庫）
	user, err := s.RepoSlave.FindUserByUsername(username)
	if err != nil {
		return "", nil, time.Time{}, errors.New("用戶不存在")
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, time.Time{}, errors.New("密碼錯誤")
	}

	// 生成 JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenString, err := s.JWTUtil.GenerateToken(user.ID, "user")
	if err != nil {
		return "", nil, time.Time{}, errors.New("生成 token 失敗")
	}

	// 轉換為 DTO
	userDTO := &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return tokenString, userDTO, expiresAt, nil
}

// GetUserByID 根據 ID 獲取用戶
func (s *UserService) GetUserByID(id uint) (*dto.UserDTO, error) {
	// 讀操作 - 使用從庫
	user, err := s.RepoSlave.FindUserByID(id)
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUser 更新用戶
func (s *UserService) UpdateUser(id uint, username string, email string, avatar string, bio string) (*dto.UserDTO, error) {
	// 讀操作 - 使用從庫獲取現有用戶數據
	user, err := s.RepoSlave.FindUserByID(id)
	if err != nil {
		return nil, err
	}

	// 更新用戶資訊
	if username != "" {
		// 檢查用戶名是否已存在（讀操作 - 使用從庫）
		if existingUser, err := s.RepoSlave.FindUserByUsername(username); err == nil && existingUser.ID != id {
			return nil, errors.New("用戶名已存在")
		}
		user.Username = username
	}

	if email != "" {
		// 檢查郵箱是否已存在（讀操作 - 使用從庫）
		if existingUser, err := s.RepoSlave.FindUserByEmail(email); err == nil && existingUser.ID != id {
			return nil, errors.New("郵箱已存在")
		}
		user.Email = email
	}

	if avatar != "" {
		user.Avatar = avatar
	}

	if bio != "" {
		user.Bio = bio
	}

	user.UpdatedAt = time.Now()

	// 寫操作 - 使用主庫
	if err := s.Repo.UpdateUser(user); err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// DeleteUser 刪除用戶
func (s *UserService) DeleteUser(id uint) error {
	// 寫操作 - 使用主庫
	return s.Repo.DeleteUser(id)
}
