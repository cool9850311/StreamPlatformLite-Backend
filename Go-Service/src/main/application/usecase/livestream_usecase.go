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
	"Go-Service/src/main/domain/interface/file_cache"
	"Go-Service/src/main/domain/interface/libarary/ffmpeg"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/util"
	"context"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LivestreamUsecase struct {
	LivestreamRepo   repository.LivestreamRepository
	Log              logger.Logger
	config           config.Config
	streamService    stream.ILivestreamService
	viewerCountCache cache.ViewerCount
	chatCache        cache.Chat
	fileCache        file_cache.IFileCache
	ffmpegLibrary    ffmpeg.FfmpegLibrary
	m3u8Lock         sync.Mutex
	convertTaskLock  sync.Mutex
}

func NewLivestreamUsecase(livestreamRepo repository.LivestreamRepository, log logger.Logger, config config.Config, streamService stream.ILivestreamService, viewerCountCache cache.ViewerCount, chatCache cache.Chat, fileCache file_cache.IFileCache, ffmpegLibrary ffmpeg.FfmpegLibrary) *LivestreamUsecase {
	u := &LivestreamUsecase{
		LivestreamRepo:   livestreamRepo,
		Log:              log,
		config:           config,
		streamService:    streamService,
		viewerCountCache: viewerCountCache,
		chatCache:        chatCache,
		fileCache:        fileCache,
		ffmpegLibrary:    ffmpegLibrary,
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

func (u *LivestreamUsecase) checkEditorRole(userRole role.Role) error {
	if userRole > role.Editor {
		return errors.ErrUnauthorized
	}
	return nil
}

// checkViewAccess 检查用户是否有权限观看直播
// 根据直播的Visibility和用户角色判断
func (u *LivestreamUsecase) checkViewAccess(userRole role.Role, visibility livestream.Visibility) error {
	switch visibility {
	case livestream.Public:
		// Public模式：所有人都可以观看（包括Anonymous）
		return nil
	case livestream.MemberOnly:
		// MemberOnly模式：需要User及以上角色（排除Anonymous和Guest）
		if userRole > role.User {
			return errors.ErrUnauthorized
		}
		return nil
	case livestream.Private:
		// Private模式：仅Admin可访问（可扩展为Owner）
		if userRole != role.Admin {
			return errors.ErrUnauthorized
		}
		return nil
	case livestream.Link:
		// Link模式：需要特定token（暂时按User及以上处理）
		if userRole > role.User {
			return errors.ErrUnauthorized
		}
		return nil
	default:
		return errors.ErrUnauthorized
	}
}

// checkChatAccess 检查用户是否有权限发送聊天
// Public直播：Guest及以上可聊天（排除Anonymous）
// MemberOnly直播：User及以上可聊天（排除Anonymous和Guest）
func (u *LivestreamUsecase) checkChatAccess(userRole role.Role, visibility livestream.Visibility) error {
	// 首先检查是否有观看权限（这会自动拦截Guest访问MemberOnly）
	if err := u.checkViewAccess(userRole, visibility); err != nil {
		return err
	}

	// Anonymous用户不能发送聊天
	if userRole == role.Anonymous {
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
		u.Log.Error(ctx, "Error getting livestream by ID: "+err.Error())
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
		IsRecord:      livestream.IsRecord,
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
		u.Log.Error(ctx, "Error getting livestream by owner ID: "+err.Error())
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
		IsRecord:      livestream.IsRecord,
	}
	return &livestreamResponse, nil
}
func (u *LivestreamUsecase) GetOne(ctx context.Context, userRole role.Role) (*livestreamDTO.LivestreamGetOneResponseDTO, error) {
	// 先获取直播信息
	livestream, err := u.LivestreamRepo.GetOne()
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream: "+err.Error())
		return nil, err
	}

	// 根据Visibility检查访问权限
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to GetOne, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
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
		StreamURL:   prefix + u.config.Server.Domain + port + "/livestream/" + livestream.UUID + "/playlist.m3u8",
		Visibility:  livestream.Visibility, // 新增字段
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
	livestreamEntity := livestream.Livestream{
		UUID:        streamUUID,
		APIKey:      apiKey,
		OwnerUserId: userID,
		Name:        livestreamData.Name,
		Visibility:  livestreamData.Visibility,
		Title:       livestreamData.Title,
		Information: livestreamData.Information,
		BanList:     []string{},
		MuteList:    []string{},
		IsRecord:    livestreamData.IsRecord,
	}
	err = u.LivestreamRepo.Create(&livestreamEntity)
	if err != nil {
		u.Log.Error(ctx, "Error creating livestream: "+err.Error())
		return nil, err
	}
	err = u.streamService.OpenStream(livestreamData.Name, streamUUID, apiKey, livestreamData.IsRecord)
	if err != nil {
		u.Log.Error(ctx, "Error opening stream Service: "+err.Error())
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
		u.Log.Error(ctx, "Error updating livestream: "+err.Error())
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
		u.Log.Error(ctx, "Error deleting livestream: "+err.Error())
		return err
	}
	err = u.streamService.CloseStream(id)
	if err != nil {
		u.Log.Error(ctx, "Error closing stream: "+err.Error())
		return err
	}
	return nil
}
func (u *LivestreamUsecase) PingViewerCount(ctx context.Context, userRole role.Role, livestreamUUID string, userID string, anonymousID string) (int, error) {
	// 获取直播信息以检查Visibility
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream: "+err.Error())
		return 0, errors.ErrNotFound
	}

	// 根据Visibility检查访问权限
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to PingViewerCount, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return 0, err
	}

	// Determine effective user ID
	effectiveUserID := userID
	if userRole == role.Anonymous {
		// Anonymous users must provide anonymousID
		if strings.TrimSpace(anonymousID) == "" {
			return 0, errors.ErrInvalidInput
		}
		effectiveUserID = anonymousID
	}

	err = u.viewerCountCache.AddViewerCount(livestreamUUID, effectiveUserID)
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
	// 获取直播信息以检查Visibility
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream: "+err.Error())
		return nil, errors.ErrNotFound
	}

	// 根据Visibility检查访问权限（观看权限即可获取聊天）
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to GetChat, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return nil, err
	}

	chats, err := u.chatCache.GetChat(livestreamUUID, index, 10)
	if err != nil {
		return nil, err
	}
	return chats, nil
}
func (u *LivestreamUsecase) AddChat(ctx context.Context, identityProvider string, userRole role.Role, livestreamUUID string, chat chat.Chat) error {
	if len(chat.Message) > 100 {
		return errors.ErrInvalidInput
	}

	// 获取直播信息以检查Visibility
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream by ID: "+err.Error())
		return err
	}

	// 根据Visibility检查聊天权限
	if err := u.checkChatAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to AddChat, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return err
	}

	if livestream.MuteList != nil && slices.Contains(livestream.MuteList, identityProvider+"-"+chat.UserID) {
		return errors.ErrMuteUser
	}
	err = u.chatCache.AddChat(livestreamUUID, chat)
	if err != nil {
		return err
	}
	return nil
}
func (u *LivestreamUsecase) DeleteChat(ctx context.Context, userRole role.Role, currentUserID string, livestreamUUID string, chatID string) error {
	// 获取直播信息以检查Visibility
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream by ID: "+err.Error())
		return err
	}

	// 根据Visibility检查访问权限（需要有观看权限才能删除聊天）
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to DeleteChat, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return err
	}

	// Editor and Admin can delete any chat
	if userRole <= role.Editor {
		// Editor can only delete User (3) and Guest (4) messages (except own messages)
		if userRole == role.Editor {
			chat, err := u.chatCache.GetChatByID(livestreamUUID, chatID)
			if err != nil {
				u.Log.Error(ctx, "Error getting chat: "+err.Error())
				return err
			}

			// Editor can always delete their own messages
			if chat.UserID != currentUserID {
				// Editor cannot delete Admin or other Editor's messages
				if chat.Role <= role.Editor {
					u.Log.Warn(ctx, "Editor cannot delete Admin or Editor's message")
					return errors.ErrUnauthorized
				}
			}
		}

		err := u.chatCache.DeleteChat(livestreamUUID, chatID)
		if err != nil {
			return err
		}
		return nil
	}

	// User and Guest can only delete their own chat
	if userRole == role.User || userRole == role.Guest {
		// Get the chat to check ownership
		chat, err := u.chatCache.GetChatByID(livestreamUUID, chatID)
		if err != nil {
			u.Log.Error(ctx, "Error getting chat: "+err.Error())
			return err
		}

		// Check if the user owns this chat
		if chat.UserID != currentUserID {
			u.Log.Error(ctx, "User attempting to delete someone else's chat")
			return errors.ErrUnauthorized
		}

		// Delete the chat
		err = u.chatCache.DeleteChat(livestreamUUID, chatID)
		if err != nil {
			return err
		}
		return nil
	}

	// Anonymous and other roles cannot delete chats
	u.Log.Error(ctx, "Unauthorized access to DeleteChat")
	return errors.ErrUnauthorized
}
func (u *LivestreamUsecase) GetDeleteChatIDs(ctx context.Context, userRole role.Role, livestreamUUID string) ([]string, error) {
	// 获取直播信息
	livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream in GetDeleteChatIDs: "+err.Error())
		return nil, errors.ErrNotFound
	}

	// 检查观看权限（与GetChat保持一致）
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to GetDeleteChatIDs, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return nil, err
	}

	ids, err := u.chatCache.GetDeleteChatIDs(livestreamUUID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
func (u *LivestreamUsecase) MuteUser(ctx context.Context, identityProvider string, userRole role.Role, currentUserID string, livestreamUUID string, chatID string) error {
	if err := u.checkEditorRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to MuteUser")
		return err
	}

	// Query real user info from chatID (same pattern as DeleteChat)
	chat, err := u.chatCache.GetChatByID(livestreamUUID, chatID)
	if err != nil {
		u.Log.Error(ctx, "Error getting chat: "+err.Error())
		return err
	}

	// Cannot mute self
	if chat.UserID == currentUserID {
		u.Log.Warn(ctx, "User attempting to mute themselves")
		return errors.ErrUnauthorized
	}

	// Editor cannot mute Admin or Editor
	if userRole == role.Editor {
		if chat.Role == role.Admin || chat.Role == role.Editor {
			u.Log.Warn(ctx, "Editor cannot mute Admin or Editor")
			return errors.ErrUnauthorized
		}
	}

	// Use real userID from queried chat
	err = u.LivestreamRepo.MuteUser(identityProvider, livestreamUUID, chat.UserID)
	if err != nil {
		return err
	}
	return nil
}
func (u *LivestreamUsecase) GetFile(ctx context.Context, filePath string, userRole role.Role) ([]byte, error) {
	// 获取直播信息以检查Visibility
	livestream, err := u.LivestreamRepo.GetOne()
	if err != nil {
		u.Log.Error(ctx, "Error getting livestream: "+err.Error())
		return nil, err
	}

	// 根据Visibility检查访问权限
	if err := u.checkViewAccess(userRole, livestream.Visibility); err != nil {
		u.Log.Warn(ctx, "Unauthorized access to GetFile, role: "+userRole.String()+", visibility: "+string(livestream.Visibility))
		return nil, err
	}

	ext := filepath.Ext(filePath)
	if ext == ".mp4" {
		u.Log.Error(ctx, "Unauthorized access to mp4")
		return nil, errors.ErrNotFound
	}
	if filepath.Base(filePath) == "record.m3u8" {
		u.Log.Error(ctx, "Unauthorized access to record.m3u8")
		return nil, errors.ErrNotFound
	}
	if ext == ".m3u8" {
		u.m3u8Lock.Lock()
		defer u.m3u8Lock.Unlock()
	}

	if data, ok := u.fileCache.LoadCache(filePath); ok {
		return data, nil
	}

	fileData, err := u.fileCache.ReadFile(filePath)
	if err != nil {
		u.Log.Error(ctx, "Error reading file: "+err.Error())
		return nil, err
	}

	u.fileCache.StoreCache(filePath, fileData)

	if ext == ".m3u8" {
		go u.updateCachePeriodically(filePath)
	}

	return fileData, nil
}
func (u *LivestreamUsecase) GetRecord(ctx context.Context, livestreamUUID string, filePath string, userRole role.Role) (string, error) {
	if err := u.checkAdminRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetRecord")
		return "", err
	}
	ext := filepath.Ext(filePath)
	if ext != ".mp4" {
		u.Log.Error(ctx, "Not mp4 file: "+filePath)
		return "", errors.ErrNotFound
	}

	fullFilePath, err := u.fileCache.GetSingleFileName(filePath)
	if err != nil {
		u.Log.Error(ctx, "Error getting single file name: "+err.Error())

		// Try to acquire the lock for conversion.
		if !u.convertTaskLock.TryLock() {
			return "", errors.ErrNotFound
		}

		// Instead of passing filePath as the source, construct the source file
		// as a "record.m3u8" file in the same directory.
		recordPath := filepath.Join(filepath.Dir(filePath), "record.m3u8")
		go func() {
			defer u.convertTaskLock.Unlock()
			u.Log.Info(ctx, "Converting stream to mp4: from "+recordPath+" to "+filePath)
			livestream, err := u.LivestreamRepo.GetByID(livestreamUUID)
			if err != nil {
				u.Log.Error(ctx, "Error getting livestream by ID: "+err.Error())
				return
			}
			err = u.ffmpegLibrary.ConvertStreamToMp4(recordPath, livestream.Title)
			if err != nil {
				u.Log.Error(ctx, "Error converting stream to mp4: "+err.Error())
			}
		}()
		return "", errors.ErrNotFound
	}
	return fullFilePath, nil
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

		u.fileCache.Range(func(key, value any) bool {
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
