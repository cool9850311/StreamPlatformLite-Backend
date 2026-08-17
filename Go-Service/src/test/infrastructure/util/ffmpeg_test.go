package infrastructure

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"Go-Service/src/main/infrastructure/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireFfmpeg skips loudly (never silently) when the real ffmpeg binary that
// these tests exercise is absent from PATH.
func requireFfmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skipf("SKIPPING: the required external binary %q was not found on PATH (%v). "+
			"These tests drive the real ffmpeg conversion pipeline and cannot be faked; "+
			"install ffmpeg to run them.", "ffmpeg", err)
	}
}

// newHLSFixture creates <root>/<recordDir>/record.m3u8 plus its segments using
// a one second lavfi test pattern, and returns the record directory and the
// m3u8 source path.
//
// The generated source is deliberately VIDEO ONLY. That exercises the
// production "-bsf:a aac_adtstoasc" flag against a stream that has no audio
// track, which is the real world shape of these recordings.
func newHLSFixture(t *testing.T, root string) (recordDir, sourcePath string) {
	t.Helper()

	recordDir = filepath.Join(root, "83636040-0000-0000-0000-000000000000")
	require.NoError(t, os.MkdirAll(recordDir, 0o755))

	cmd := exec.Command(
		"ffmpeg",
		"-loglevel", "error",
		"-f", "lavfi",
		"-i", "testsrc=duration=1:size=64x64:rate=5",
		"-c:v", "libx264",
		"-f", "hls",
		"-hls_time", "1",
		"-hls_list_size", "0",
		"-hls_segment_filename", "seg%d.ts",
		"record.m3u8",
	)
	cmd.Dir = recordDir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "failed to generate the HLS fixture: %s", out)

	sourcePath = filepath.Join(recordDir, "record.m3u8")
	require.FileExists(t, sourcePath)
	return recordDir, sourcePath
}

// snapshotTree lists every path under root, relative to root, excluding
// anything inside skipDir. It is used to prove the converter created no file
// outside the record directory.
func snapshotTree(t *testing.T, root, skipDir string) []string {
	t.Helper()

	var paths []string
	require.NoError(t, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == skipDir {
			return filepath.SkipDir
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		paths = append(paths, rel)
		return nil
	}))
	sort.Strings(paths)
	return paths
}

// B-1: a successful conversion publishes exactly one file, under the fixed
// constant name, inside the record directory and nowhere else.
func TestConvertStreamToMp4_PublishesFixedName(t *testing.T) {
	requireFfmpeg(t)

	root := t.TempDir()
	recordDir, sourcePath := newHLSFixture(t, root)

	before := snapshotTree(t, root, recordDir)

	require.NoError(t, util.NewFfmpegLibrary().ConvertStreamToMp4(sourcePath))

	assert.FileExists(t, filepath.Join(recordDir, util.RecordingFileName))

	matches, err := filepath.Glob(filepath.Join(recordDir, "*.mp4"))
	require.NoError(t, err)
	assert.Len(t, matches, 1, "the readiness glob must see exactly one mp4, got %v", matches)
	if len(matches) == 1 {
		assert.Equal(t, util.RecordingFileName, filepath.Base(matches[0]))
	}

	after := snapshotTree(t, root, recordDir)
	assert.Equal(t, before, after, "no file may be created outside the record directory")
}

// B-2: a failed conversion publishes nothing and leaves no partial behind.
func TestConvertStreamToMp4_FailureDoesNotPublish(t *testing.T) {
	requireFfmpeg(t)

	recordDir := t.TempDir()
	sourcePath := filepath.Join(recordDir, "does-not-exist.m3u8")

	err := util.NewFfmpegLibrary().ConvertStreamToMp4(sourcePath)

	assert.Error(t, err, "converting a non-existent source must fail")
	assert.NoFileExists(t, filepath.Join(recordDir, util.RecordingFileName),
		"a failed conversion must never publish the final recording")
	assert.NoFileExists(t, filepath.Join(recordDir, util.RecordingPartialName),
		"a failed conversion must clean up its partial file")
}

// B-3: the partial name must be invisible to the readiness glob, verified with
// real glob semantics rather than a string comparison.
func TestConvertStreamToMp4_PartialIsNotGlobbable(t *testing.T) {
	matched, err := filepath.Match("*.mp4", util.RecordingPartialName)
	require.NoError(t, err)
	assert.False(t, matched,
		"RecordingPartialName %q must not match the readiness glob *.mp4, "+
			"otherwise a half-written conversion would be served as a finished recording",
		util.RecordingPartialName)

	matched, err = filepath.Match("*.mp4", util.RecordingFileName)
	require.NoError(t, err)
	assert.True(t, matched, "RecordingFileName %q must match the readiness glob *.mp4", util.RecordingFileName)
}

// B-4 (adapted): a partial left behind by a crashed run must not wedge the
// feature. Since the production command has no "-y", ffmpeg would refuse to
// overwrite the leftover and every retry would fail forever; the converter
// clears the stale partial before starting.
func TestConvertStreamToMp4_StalePartialIsRetried(t *testing.T) {
	requireFfmpeg(t)

	root := t.TempDir()
	recordDir, sourcePath := newHLSFixture(t, root)

	stalePartial := filepath.Join(recordDir, util.RecordingPartialName)
	require.NoError(t, os.WriteFile(stalePartial, []byte("junk left by a crashed run"), 0o644))

	require.NoError(t, util.NewFfmpegLibrary().ConvertStreamToMp4(sourcePath),
		"a stale partial must not block a retry")

	assert.FileExists(t, filepath.Join(recordDir, util.RecordingFileName))
	assert.NoFileExists(t, stalePartial, "the partial must be consumed by the atomic publish")

	info, err := os.Stat(filepath.Join(recordDir, util.RecordingFileName))
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(len("junk left by a crashed run")),
		"the published recording must be the freshly converted file, not the stale junk")
}
