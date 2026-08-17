package ffmpeg

type FfmpegLibrary interface {
	// ConvertStreamToMp4 converts the stream at sourcePath into an MP4 written
	// next to it, under an implementation-owned constant filename.
	//
	// The output filename is intentionally NOT a parameter: it used to be derived
	// from the user-supplied livestream Title, which allowed path traversal,
	// absolute-path escape and ffmpeg argument injection. Callers must not be able
	// to influence the output path.
	ConvertStreamToMp4(sourcePath string) error
}
