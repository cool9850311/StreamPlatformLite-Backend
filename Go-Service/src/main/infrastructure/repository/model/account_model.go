package model

type AccountModel struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     int8   `gorm:"not null"`
}

func (AccountModel) TableName() string { return "accounts" }
