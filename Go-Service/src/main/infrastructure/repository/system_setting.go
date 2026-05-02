package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/system"
	"Go-Service/src/main/infrastructure/repository/model"
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type PostgresSystemSettingRepository struct {
	db *gorm.DB
}

func NewPostgresSystemSettingRepository(db *gorm.DB) repository.SystemSettingRepository {
	return &PostgresSystemSettingRepository{db: db}
}

func (r *PostgresSystemSettingRepository) GetSetting() (*system.Setting, error) {
	var m model.SystemSettingModel
	result := r.db.First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return &system.Setting{
			EditorRoleId:        "",
			StreamAccessRoleIds: []string{},
		}, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &system.Setting{
		EditorRoleId:        m.EditorRoleID,
		StreamAccessRoleIds: []string(m.StreamAccessRoleIDs),
	}, nil
}

func (r *PostgresSystemSettingRepository) SetSetting(setting *system.Setting) error {
	m := model.SystemSettingModel{
		ID:                  1,
		EditorRoleID:        setting.EditorRoleId,
		StreamAccessRoleIDs: pq.StringArray(setting.StreamAccessRoleIds),
	}
	return r.db.Save(&m).Error
}
