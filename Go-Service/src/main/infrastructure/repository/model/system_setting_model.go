package model

import "github.com/lib/pq"

type SystemSettingModel struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement"`
	EditorRoleID        string         `gorm:"column:editor_role_id;not null;default:''"`
	StreamAccessRoleIDs pq.StringArray `gorm:"type:text[];column:stream_access_role_ids;not null;default:'{}'"`
}

func (SystemSettingModel) TableName() string { return "system_settings" }
