package model

import "github.com/lib/pq"

type LivestreamModel struct {
	UUID        string         `gorm:"primaryKey"`
	Name        string         `gorm:"not null"`
	APIKey      string         `gorm:"column:api_key;not null"`
	OwnerUserID string         `gorm:"column:owner_user_id;not null"`
	Visibility  string         `gorm:"not null"`
	Title       string         `gorm:"not null;default:''"`
	Information string         `gorm:"not null;default:''"`
	BanList     pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	MuteList    pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	IsRecord    bool           `gorm:"column:is_record;not null;default:false"`
}

func (LivestreamModel) TableName() string { return "livestreams" }
