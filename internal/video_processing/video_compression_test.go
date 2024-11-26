package video_processing

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReturnValidCompressionRates(t *testing.T) {
	assert.DeepEqual(t,
		ReturnValidCompressionRates(720),
		[]string{"426x240", "640x360", "854x480"})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(721),
		[]string{"426x240", "640x360", "854x480", "1280x720"})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(2160),
		[]string{"426x240", "640x360", "854x480", "1280x720", "1920x1080", "2560x1440"})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(100000000),
		[]string{"426x240", "640x360", "854x480", "1280x720", "1920x1080", "2560x1440", "3840x2160"})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(0),
		[]string{})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(-11),
		[]string{})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(144),
		[]string{})
	assert.DeepEqual(t,
		ReturnValidCompressionRates(240),
		[]string{})
}

func TestChangeCodec(t *testing.T) {
	inFileName := "test/data/Horses.mp4"
	outFileName := "test/data/Horses_240p.mp4"

	ChangeCodec(inFileName, outFileName, "426x240")
	w, h := GetVideoSize(outFileName)
	assert.Equal(t, w, 426)
	assert.Equal(t, h, 240)
	err := os.Remove(outFileName)
	if err != nil {
		panic(err)
	}
}

func TestChunkVideo(t *testing.T) {
	inFileName := []string{
		"../../test/data/horses_compressed/Horses_240p.mp4",
		"../../test/data/horses_compressed/Horses_360p.mp4",
		"../../test/data/horses_compressed/Horses_480p.mp4",
		"../../test/data/horses_compressed/Horses_720p.mp4"}

	ChunkVideo(inFileName, 8)
}
