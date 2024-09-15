package repository

import (
	"Go-Service/src/main/domain/entity/system"
)

type SystemSettingRepository interface {
	GetSetting() (*system.Setting, error)
}
