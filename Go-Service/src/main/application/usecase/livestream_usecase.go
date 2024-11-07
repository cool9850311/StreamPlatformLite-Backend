package usecase

import (
	"Go-Service/src/main/application/dto/config"
	livestreamDTO "Go-Service/src/main/application/dto/livestream"
	"Go-Service/src/main/application/interface/cache"
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/domain/entity/chat"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/util"
	"context"
	"strconv"
	"Go-Service/src/main/domain/interface/file_cache"
	"github.com/google/uuid"
	"sync"
	"path/filepath"
	"time"
	"strings"
)

type LivestreamUsecase struct {
	LivestreamRepo   repository.LivestreamRepository
	Log              logger.Logger
	config           config.Config
	streamService    stream.ILivestreamService
	viewerCountCache cache.ViewerCount
	chatCache        cache.Chat
	fileCache       file_cache.IFileCache
	m3u8Lock        sync.Mutex
}

func NewLivestreamUsecase(livestreamRepo repository.LivestreamRepository, log logger.Logger, config config.Config, streamService stream.ILivestreamService, viewerCountCache cache.ViewerCount, chatCache cache.Chat, fileCache file_cache.IFileCache) *LivestreamUsecase {
	u := &LivestreamUsecase{
		LivestreamRepo:   livestreamRepo,
		Log:              log,
		config:           config,
		streamService:    streamService,
		viewerCountCache: viewerCountCache,
		chatCache:        chatCache,
		fileCache:        fileCache,
	}
	go u.startCacheCleanup()
	return u
}

func (u *LivestreamUsecase) checkAdminRole(userRole role.Role) error {
	if userRole != role.Admin {
		return errors.ErrUnauthorized
	}
	return nil
}

func (u *LivestreamUsecase) checkUserRole(userRole role.Role) error {
	if userRole > role.User {
		return errors.ErrUnauthorized
	}
	return nil
}
func (u *LivestreamUsecase) checkEditorRole(userRole role.Role) error {
	if userRole > role.Editor {
		return errors.ErrUnauthorized
	}
	return nil
}

func (u *LivestreamUsecase) GetLivestreamByID(ctx context.Context, id string, userRole role.Role) (*livestreamDTO.LivestreamGetByOwnerIDResponseDTO, error) {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetLivestreamByID")
		return nil, err
	}
	livestream, err := u.LivestreamRepo.GetByID(id)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream by ID")
		return nil, err
	}
	livestreamResponse := livestreamDTO.LivestreamGetByOwnerIDResponseDTO{
		UUID:          livestream.UUID,
		Name:          livestream.Name,
		Visibility:    livestream.Visibility,
		Title:         livestream.Title,
		Information:   livestream.Information,
		StreamPushURL: "rtmp://" + u.config.Server.Domain + ":1935/" + livestream.APIKey,
		BanList:       livestream.BanList,
		MuteList:      livestream.MuteList,
	}
	return &livestreamResponse, nil
}

func (u *LivestreamUsecase) GetLivestreamByOwnerID(ctx context.Context, ownerID string, userRole role.Role) (*livestreamDTO.LivestreamGetByOwnerIDResponseDTO, error) {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetLivestreamByOwnerID")
		return nil, err
	}
	livestream, err := u.LivestreamRepo.GetByOwnerID(ownerID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream by owner ID")
		return nil, err
	}
	livestreamResponse := livestreamDTO.LivestreamGetByOwnerIDResponseDTO{
		UUID:          livestream.UUID,
		Name:          livestream.Name,
		Visibility:    livestream.Visibility,
		Title:         livestream.Title,
		Information:   livestream.Information,
		StreamPushURL: "rtmp://" + u.config.Server.Domain + ":1935/" + livestream.APIKey,
		BanList:       livestream.BanList,
		MuteList:      livestream.MuteList,
	}
	return &livestreamResponse, nil
}
func (u *LivestreamUsecase) GetOne(ctx context.Context, userRole role.Role) (*livestreamDTO.LivestreamGetOneResponseDTO, error) {
	if err := u.checkUserRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetOne")
		return nil, err
	}
	livestream, err := u.LivestreamRepo.GetOne()
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream")
		return nil, err
	}
	prefix := "http://"
	port := ":" + strconv.Itoa(u.config.Server.Port)
	if u.config.Server.HTTPS {
		prefix = "https://"
		port = ""
	}

	livestreamResponse := livestreamDTO.LivestreamGetOneResponseDTO{
		UUID:        livestream.UUID,
		Name:        livestream.Name,
		Title:       livestream.Title,
		Information: livestream.Information,
		StreamURL:   prefix + u.config.Server.Domain + port + "/livestream/" + livestream.OutputPathUUID + "/playlist.m3u8",
	}
	return &livestreamResponse, nil
}

func (u *LivestreamUsecase) CreateLivestream(ctx context.Context, livestreamData *livestreamDTO.LivestreamCreateDTO, userID string, userRole role.Role) (*livestreamDTO.LivestreamCreateResponseDTO, error) {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to CreateLivestream")
		return nil, err
	}
	_, err := u.LivestreamRepo.GetOne()
	if err == nil {
		u.Log.Error(ctx, "Livestream already exists")
		return nil, errors.ErrExists
	}
	apiKey, err := util.GenerateRandomBase64String(16)
	if err != nil {
		u.Log.Error(ctx, "Error generating API key")
		return nil, err
	}
	streamUUID := uuid.New().String()
	outputPathUUID := uuid.New().String()
	livestreamEntity := livestream.Livestream{
		UUID:           streamUUID,
		APIKey:         apiKey,
		OutputPathUUID: outputPathUUID,
		OwnerUserId:    userID,
		Name:           livestreamData.Name,
		Visibility:     livestreamData.Visibility,
		Title:          livestreamData.Title,
		Information:    livestreamData.Information,
		BanList:        []string{},
		MuteList:       []string{},
	}
	err = u.LivestreamRepo.Create(&livestreamEntity)
	if err != nil {
		u.Log.Error(ctx, "Error creating livestream")
		return nil, err
	}
	err = u.streamService.OpenStream(livestreamData.Name, streamUUID, apiKey, outputPathUUID)
	if err != nil {
		u.Log.Error(ctx, "Error opening stream Service")
		return nil, err
	}
	return &livestreamDTO.LivestreamCreateResponseDTO{
		StreamPushURL: "rtmp://" + u.config.Server.Domain + ":1935/" + apiKey,
	}, nil
}

func (u *LivestreamUsecase) UpdateLivestream(ctx context.Context, livestream *livestream.Livestream, userRole role.Role) error {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to UpdateLivestream")
		return err
	}
	err := u.LivestreamRepo.Update(livestream)
	if err != nil {
		u.Log.Error(ctx, "Error updating livestream")
		return err
	}
	return nil
}

func (u *LivestreamUsecase) DeleteLivestream(ctx context.Context, id string, userRole role.Role) error {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to DeleteLivestream")
		return err
	}
	err := u.LivestreamRepo.Delete(id)
	if err != nil {
		u.Log.Error(ctx, "Error deleting livestream")
		return err
	}
	err = u.streamService.CloseStream(id)
	if err != nil {
		u.Log.Error(ctx, "Error closing stream")
		return err
	}
	return nil
}
func (u *LivestreamUsecase) PingViewerCount(ctx context.Context, userRole role.Role, livestreamUUID string, userID string) (int, error) {
	if err := u.checkUserRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to PingViewerCount")
		return 0, err
	}
	err := u.viewerCountCache.AddViewerCount(livestreamUUID, userID)
	if err != nil {
		return 0, err
	}
	viewerCount, err := u.viewerCountCache.GetViewerCount(livestreamUUID)
	if err != nil {
		return 0, err
	}
	return viewerCount, nil
}

// remove every viewer count that is older than 5 seconds cron job
func (u *LivestreamUsecase) RemoveViewerCount(ctx context.Context, livestreamUUID string, seconds int) (int, error) {
	viewerCount, err := u.viewerCountCache.RemoveViewerCount(livestreamUUID, seconds)
	if err != nil {
		return 0, err
	}
	return viewerCount, nil
}
func (u *LivestreamUsecase) GetChat(ctx context.Context, userRole role.Role, livestreamUUID string, index string) ([]chat.Chat, error) {
	if err := u.checkUserRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetChat")
		return nil, err
	}
	chats, err := u.chatCache.GetChat(livestreamUUID, index, 10)
	if err != nil {
		return nil, err
	}
	return chats, nil
}
func (u *LivestreamUsecase) AddChat(ctx context.Context, userRole role.Role, livestreamUUID string, chat chat.Chat) error {
	if err := u.checkUserRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to AddChat")
		return err
	}
	if len(chat.Message) > 100 {
		return errors.ErrInvalidInput
	}
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream by ID")
		return err
	}
	if livestream.MuteList != nil {
		for _, userID := range livestream.MuteList {
			if userID == chat.UserID {
				return errors.ErrMuteUser
			}
		}
	}
	err = u.chatCache.AddChat(livestreamUUID, chat)
	if err != nil {
		return err
	}
	return nil
}
func (u *LivestreamUsecase) DeleteChat(ctx context.Context, userRole role.Role, livestreamUUID string, chatID string) error {
	if err := u.checkEditorRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to DeleteChat")
		return err
	}
	err := u.chatCache.DeleteChat(livestreamUUID, chatID)
	if err != nil {
		return err
	}
	return nil
}
func (u *LivestreamUsecase) GetDeleteChatIDs(ctx context.Context, userRole role.Role, livestreamUUID string) ([]string, error) {
	if err := u.checkUserRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetDeleteChatIDs")
		return nil, err
	}
	ids, err := u.chatCache.GetDeleteChatIDs(livestreamUUID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
func (u *LivestreamUsecase) MuteUser(ctx context.Context, userRole role.Role, livestreamUUID string, userID string) error {
	if err := u.checkEditorRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to MuteUser")
		return err
	}
	err := u.LivestreamRepo.MuteUser(livestreamUUID, userID)
	if err != nil {
		return err
	}
	return nil
}
func (u *LivestreamUsecase) GetFile(filePath string) ([]byte, error) {
	ext := filepath.Ext(filePath)

	if ext == ".m3u8" {
		u.m3u8Lock.Lock()
		defer u.m3u8Lock.Unlock()
	}

	if data, ok := u.fileCache.LoadCache(filePath); ok {
		return data, nil
	}

	fileData, err := u.fileCache.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	u.fileCache.StoreCache(filePath, fileData)

	if ext == ".m3u8" {
		go u.updateCachePeriodically(filePath)
	}

	return fileData, nil
}
func (u *LivestreamUsecase) updateCachePeriodically(filePath string) {
	for {
		time.Sleep(1 * time.Second)
		u.m3u8Lock.Lock()
		fileData, err := u.fileCache.ReadFile(filePath)
		if err == nil {
			u.fileCache.StoreCache(filePath, fileData)
		}
		u.m3u8Lock.Unlock()
	}
}

func (u *LivestreamUsecase) startCacheCleanup() {
	for {
		time.Sleep(10 * time.Second) // Run cleanup every 10 seconds
		now := time.Now().UnixMilli()

		u.fileCache.Range(func(key, value interface{}) bool {
			filePath := key.(string)
			filename := filepath.Base(filePath)

			// Extract timestamp from filename
			parts := strings.Split(filename, "-")
			if len(parts) < 3 {
				return true // Continue to next item
			}

			timestampStr := parts[1]
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				return true // Continue to next item
			}

			// Check if the file is older than 30 seconds
			if now-timestamp > 30000 {
				u.fileCache.DeleteFile(filePath)
			}

			return true // Continue to next item
		})
	}
}