package video_processing

import (
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VideoCompressionSpec struct {
	VideoPath          string
	CompressionQuality string
	ChunkSegments      int
}

func ChangeCodec(inputVideo, outputVideo, compressionQuality string) {
	err := ffmpeg.Input(inputVideo, nil).
		Output(outputVideo, ffmpeg.KwArgs{
			"s":   compressionQuality,
			"c:a": "copy",
		}).
		OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		panic(err)
	}
}

func RunStream(inFile, outFile string) {
	w, h := GetVideoSize(inFile)
	log.Println(w, h)
	validCompressionResolutions := ReturnValidCompressionRates(h)
	log.Println(validCompressionResolutions)
	ChangeCodec(inFile, outFile, validCompressionResolutions[1])
	log.Println("Done")
}
