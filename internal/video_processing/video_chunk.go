package video_processing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func chunkVideo(inputVideo []string, chunkSegments int) {
	if chunkSegments <= 0 {
		chunkSegments = 3
	}
	basePath := filepath.Dir(inputVideo[0])
	//extension := filepath.Ext(inputVideo)

	sum := sha256.Sum256([]byte(inputVideo[0] + time.Now().String()))
	sumHex := hex.EncodeToString(sum[:])
	abs, err := filepath.Abs(basePath)
	if err != nil {
		panic(err)
	}
	fullPath := abs + "/" + sumHex
	fmt.Println("creating new folder:", fullPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err := os.Mkdir(fullPath, 0755)
		if err != nil {
			panic(err)
		}
	}
	for _, video := range inputVideo {
		fmt.Println("processing video:", video)
		videoName := strings.Split(video, "/")[len(strings.Split(video, "/"))-1]
		// ffmpeg -i input.mp4 -codec: copy -start_number 0 -hls_time 10 -hls_list_size 0 -f hls playlist.m3u8
		err = ffmpeg.Input(video).
			Output(fullPath+"/playlist_"+videoName+".m3u8", ffmpeg.KwArgs{
				"codec":         "copy",
				"start_number":  "0",
				"hls_time":      strconv.Itoa(chunkSegments),
				"hls_list_size": "0",
				"f":             "hls",
			}).
			OverWriteOutput().ErrorToStdOut().Run()
		if err != nil {
			panic(err)
		}
	}
}
