package repository

import (
	"gost-panel/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户仓库
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// FindByUsername 根据用户名查询用户
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据 ID 查询用户
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.DB.Save(user).Error
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint) error {
	return r.DB.Delete(&model.User{}, id).Error
}

// List 查询用户列表
func (r *UserRepository) List(opt *QueryOption) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := r.DB.Model(&model.User{})

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用查询选项
	db = ApplyOptions(db, opt)

	if err := db.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
