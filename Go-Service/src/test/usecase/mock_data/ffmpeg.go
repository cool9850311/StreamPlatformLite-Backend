package mock_data

import "Go-Service/src/main/domain/interface/libarary/ffmpeg"

type MockFfmpegLibrary struct {
	ffmpeg.FfmpegLibrary
}

func (m *MockFfmpegLibrary) ConvertStreamToMp4(filePath string, fileName string) error {
	return nil
}
