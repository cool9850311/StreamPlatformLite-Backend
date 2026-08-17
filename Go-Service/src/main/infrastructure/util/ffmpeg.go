package util

import (
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// RecordingFileName is the compile-time constant name of the published
	// recording. It is deliberately NOT derived from any user-supplied value
	// (previously the livestream Title was concatenated with ".mp4").
	//
	// Making the output name a constant structurally eliminates three bug
	// classes that were empirically reproduced against the old code:
	//   - Path traversal:      Title = "../../pwned" wrote outside the uuid dir.
	//   - Absolute path escape: Title = "/tmp/x" ignored cmd.Dir entirely.
	//   - Argument injection:  Title = "-y" was parsed by ffmpeg as an option,
	//                          permanently breaking conversion (DoS).
	// User data now never reaches the exec argument vector at all.
	RecordingFileName = "recording.mp4"

	// RecordingPartialName is the in-progress temp file that ffmpeg writes to
	// before the atomic publish.
	//
	// CRITICAL: this name must never match the glob "*.mp4". The usecase globs
	// "*.mp4" to decide whether a recording is ready for download; if the
	// partial file matched, a half-written (or crashed) conversion would be
	// reported as ready and served to users as a complete recording.
	// ".recording.partial" has no ".mp4" suffix, so the glob cannot see it.
	RecordingPartialName = ".recording.partial"
)

type FfmpegLibrary struct{}

func NewFfmpegLibrary() *FfmpegLibrary {
	return &FfmpegLibrary{}
}

// ConvertStreamToMp4 remuxes the HLS stream at sourcePath into an MP4 published
// as RecordingFileName inside the source directory.
//
// It takes only the source path: the output name is a constant (see
// RecordingFileName), so no caller-supplied string can influence where ffmpeg
// writes or how ffmpeg parses its command line.
//
// Atomic publish rationale:
// ffmpeg writes to RecordingPartialName, and only a successful exit is followed
// by os.Rename to RecordingFileName. Rename within a single directory is atomic,
// so RecordingFileName either does not exist or is a fully written file — it is
// never observable in a half-written state. This also fixes a pre-existing latent
// bug: a conversion that crashed mid-way used to leave a truncated "*.mp4" behind,
// which the readiness glob happily reported as a finished recording.
//
// No "-y" flag is used. Overwrite prompting was the one thing preventing the old
// code from clobbering arbitrary files, and the fixed output name makes "-y"
// unnecessary: the temp file is removed before every run, so ffmpeg always writes
// to a fresh path and never has a reason to prompt.
func (f *FfmpegLibrary) ConvertStreamToMp4(sourcePath string) error {
	sourceDir := filepath.Dir(sourcePath)

	outputPath := filepath.Join(sourceDir, RecordingFileName)
	partialPath := filepath.Join(sourceDir, RecordingPartialName)

	// Defense in depth. The names are constants so this cannot fail in practice,
	// but the assertion guarantees that any future change which reintroduces a
	// dynamic output name fails closed instead of writing outside the directory.
	// ffmpeg is not started when the boundary check fails.
	if err := ValidatePathBoundary(outputPath, sourceDir); err != nil {
		return err
	}
	if err := ValidatePathBoundary(partialPath, sourceDir); err != nil {
		return err
	}

	// Remove any stale partial left behind by a previously crashed or killed run.
	// Without "-y" and with stdin at /dev/null, ffmpeg refuses to overwrite an
	// existing output and exits with an error instead of prompting, so a leftover
	// partial would make every retry fail forever until someone deleted the file
	// by hand. Clearing it here is what makes retry-after-crash work.
	if err := os.Remove(partialPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// The output is passed as a "./"-prefixed relative path so it can never be
	// interpreted as an option, and is resolved against cmd.Dir.
	//
	// "-f mp4" is required, not optional: ffmpeg normally picks the muxer from
	// the output extension, and RecordingPartialName deliberately has none
	// (so the "*.mp4" readiness glob cannot see it). Without an explicit format
	// ffmpeg aborts with "Unable to choose an output format". The transcode
	// settings themselves are unchanged ("-c copy -bsf:a aac_adtstoasc"), and
	// mp4 is exactly the container the old ".mp4" extension selected implicitly.
	cmd := exec.Command(
		"ffmpeg",
		"-i", sourcePath,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "mp4",
		"./"+RecordingPartialName,
	)
	cmd.Dir = sourceDir

	if err := cmd.Run(); err != nil {
		// Best-effort cleanup: never publish, and never leave a partial that
		// would block the next attempt.
		_ = os.Remove(partialPath)
		return err
	}

	if err := os.Rename(partialPath, outputPath); err != nil {
		_ = os.Remove(partialPath)
		return err
	}

	return nil
}
