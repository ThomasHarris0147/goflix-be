package video_processing

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetVideoSize(fileName string) (int, int) {
	log.Println("Getting video size for", fileName)
	data, err := ffmpeg.Probe(fileName)
	if err != nil {
		panic(err)
	}
	log.Println("got video info", data)
	type VideoInfo struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int
			Height    int
		} `json:"streams"`
	}
	vInfo := &VideoInfo{}
	err = json.Unmarshal([]byte(data), vInfo)
	if err != nil {
		panic(err)
	}
	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			return s.Width, s.Height
		}
	}
	return 0, 0
}

func ReturnValidCompressionRates(resolution int) []string {
	allValidResolutions := []string{
		"426x240",
		"640x360",
		"854x480",
		"1280x720",
		"1920x1080",
		"2560x1440",
		"3840x2160",
	}
	if resolution > 2160 {
		return allValidResolutions
	}
	for index, resString := range allValidResolutions {
		resHeight := strings.Split(resString, "x")
		resInt, err := strconv.Atoi(resHeight[1])
		if err != nil {
			panic(err)
		}
		if resolution <= resInt {
			return allValidResolutions[:index]
		}
	}
	return []string{}
}
