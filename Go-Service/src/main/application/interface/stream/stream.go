package stream

type ILivestreamService interface {
	OpenStream(name, uuid, apiKey string) error
	CloseStream(uuid string) error
	StartService() error
	RunLoop() error
	IsLiveStreamExist(uuid string) bool
}
