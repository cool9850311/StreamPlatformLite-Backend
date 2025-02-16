package stream

type ILivestreamService interface {
	OpenStream(name, uuid, apiKey string, isRecord bool) error
	CloseStream(uuid string) error
	StartService() error
	RunLoop() error
	IsLiveStreamExist(uuid string) bool
}
