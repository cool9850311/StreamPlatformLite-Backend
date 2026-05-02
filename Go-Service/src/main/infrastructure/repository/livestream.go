package repository

import (
	"Go-Service/src/main/application/interface/repository"
	domainErrors "Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/infrastructure/repository/model"
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type PostgresLivestreamRepository struct {
	db *gorm.DB
}

func NewPostgresLivestreamRepository(db *gorm.DB) repository.LivestreamRepository {
	return &PostgresLivestreamRepository{db: db}
}

func toLivestreamEntity(m model.LivestreamModel) *livestream.Livestream {
	return &livestream.Livestream{
		UUID:        m.UUID,
		Name:        m.Name,
		APIKey:      m.APIKey,
		OwnerUserId: m.OwnerUserID,
		Visibility:  livestream.Visibility(m.Visibility),
		Title:       m.Title,
		Information: m.Information,
		BanList:     []string(m.BanList),
		MuteList:    []string(m.MuteList),
		IsRecord:    m.IsRecord,
	}
}

func toModel(ls *livestream.Livestream) model.LivestreamModel {
	banList := ls.BanList
	if banList == nil {
		banList = []string{}
	}
	muteList := ls.MuteList
	if muteList == nil {
		muteList = []string{}
	}
	return model.LivestreamModel{
		UUID:        ls.UUID,
		Name:        ls.Name,
		APIKey:      ls.APIKey,
		OwnerUserID: ls.OwnerUserId,
		Visibility:  string(ls.Visibility),
		Title:       ls.Title,
		Information: ls.Information,
		BanList:     pq.StringArray(banList),
		MuteList:    pq.StringArray(muteList),
		IsRecord:    ls.IsRecord,
	}
}

func (r *PostgresLivestreamRepository) GetByID(id string) (*livestream.Livestream, error) {
	var m model.LivestreamModel
	result := r.db.Where("uuid = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainErrors.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toLivestreamEntity(m), nil
}

func (r *PostgresLivestreamRepository) GetByOwnerID(ownerID string) (*livestream.Livestream, error) {
	var m model.LivestreamModel
	result := r.db.Where("owner_user_id = ?", ownerID).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainErrors.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toLivestreamEntity(m), nil
}

func (r *PostgresLivestreamRepository) GetOne() (*livestream.Livestream, error) {
	var m model.LivestreamModel
	result := r.db.First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainErrors.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toLivestreamEntity(m), nil
}

func (r *PostgresLivestreamRepository) Create(ls *livestream.Livestream) error {
	m := toModel(ls)
	return r.db.Create(&m).Error
}

func (r *PostgresLivestreamRepository) Update(ls *livestream.Livestream) error {
	m := toModel(ls)
	return r.db.Where("uuid = ?", ls.UUID).Save(&m).Error
}

func (r *PostgresLivestreamRepository) Delete(id string) error {
	return r.db.Where("uuid = ?", id).Delete(&model.LivestreamModel{}).Error
}

func (r *PostgresLivestreamRepository) MuteUser(identityProvider string, livestreamUUID string, userID string) error {
	muteEntry := identityProvider + "-" + userID
	return r.db.Exec(
		"UPDATE livestreams SET mute_list = array_append(mute_list, ?) WHERE uuid = ? AND NOT (? = ANY(mute_list))",
		muteEntry, livestreamUUID, muteEntry,
	).Error
}
