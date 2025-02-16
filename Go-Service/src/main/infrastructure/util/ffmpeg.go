package util

import (
	"os/exec"
	"path/filepath"
)

type FfmpegLibrary struct{}

func NewFfmpegLibrary() *FfmpegLibrary {
	return &FfmpegLibrary{}
}

func (f *FfmpegLibrary) ConvertStreamToMp4(filePath string, fileName string) error {
	// Get directory path from filePath
	dir := filepath.Dir(filePath)

	// Create command and set working directory
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-bsf:a", "aac_adtstoasc", fileName+".mp4")
	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
