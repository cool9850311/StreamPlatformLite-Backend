package ffmpeg

type FfmpegLibrary interface {
	ConvertStreamToMp4(filePath string, fileName string) error
}
