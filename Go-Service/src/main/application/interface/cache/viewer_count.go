package cache

type ViewerCount interface {
	GetViewerCount(livestreamUUID string) (int, error)
	AddViewerCount(livestreamUUID string, userID string) error
	RemoveViewerCount(livestreamUUID string, seconds int) (int, error)
}
