package controllers

import (
	"github.com/julienschmidt/httprouter"
	"io"
	"jameesjohn.com/uploader-api/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func WatchVideo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	videoID := p.ByName("videoID")

	videoModel := GetVideoModel()
	videoData, err := videoModel.FindVideoByID(videoID)
	if err != nil {
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	if !videoData.UploadComplete {
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": "Video upload not completed"})
		return
	}

	rootDir := getVideoRootDir(videoID)

	finalPath := filepath.Join(rootDir, "final")
	videoFile, err := os.Open(finalPath)
	if err != nil {
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}
	defer videoFile.Close()

	chunkSize := 1024 * 1024 // 1MB
	buffer := make([]byte, chunkSize)

	_, err = videoFile.Stat()
	if err != nil {
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}

	//w.Header().Set("Content-Length", fmt.Sprintf("%v", stats.Size()))

	for {
		log.Println("looping")
		_, err := videoFile.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		if err == io.EOF {
			break
		}

		bytesRead, err := w.Write(buffer)
		log.Println("Written", bytesRead, "bytes")
		if err != nil {
			utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
			return
		}
	}

}
