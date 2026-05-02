package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/account"
	domainErrors "Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/infrastructure/repository/model"
	"errors"
	"gorm.io/gorm"
)

type PostgresAccountRepository struct {
	db *gorm.DB
}

func NewPostgresAccountRepository(db *gorm.DB) repository.AccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Create(acc account.Account) error {
	m := model.AccountModel{
		Username: acc.Username,
		Password: acc.Password,
		Role:     int8(acc.Role),
	}
	return r.db.Create(&m).Error
}

func (r *PostgresAccountRepository) GetAll() ([]account.Account, error) {
	var models []model.AccountModel
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	accounts := make([]account.Account, len(models))
	for i, m := range models {
		accounts[i] = account.Account{
			ID:       m.ID,
			Username: m.Username,
			Password: m.Password,
			Role:     role.Role(m.Role),
		}
	}
	return accounts, nil
}

func (r *PostgresAccountRepository) GetByUsername(username string) (*account.Account, error) {
	var m model.AccountModel
	result := r.db.Where("username = ?", username).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainErrors.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &account.Account{
		ID:       m.ID,
		Username: m.Username,
		Password: m.Password,
		Role:     role.Role(m.Role),
	}, nil
}

func (r *PostgresAccountRepository) Update(acc account.Account) error {
	return r.db.Model(&model.AccountModel{}).Where("username = ?", acc.Username).Updates(map[string]interface{}{
		"password": acc.Password,
		"role":     int8(acc.Role),
	}).Error
}

func (r *PostgresAccountRepository) Delete(username string) error {
	return r.db.Where("username = ?", username).Delete(&model.AccountModel{}).Error
}
