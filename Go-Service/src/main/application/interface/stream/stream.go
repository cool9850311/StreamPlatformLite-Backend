package stream

type LivestreamServiceInterface interface {
	OpenStream(name, uuid, apiKey string) error
	CloseStream(uuid string) error
}
