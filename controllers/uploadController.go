package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"io/fs"
	"jameesjohn.com/uploader-api/config"
	"jameesjohn.com/uploader-api/database"
	"jameesjohn.com/uploader-api/models"
	"jameesjohn.com/uploader-api/utils"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func GetVideoModel() models.VideoModel {
	vModel := models.VideoModel{Db: database.Db}

	return vModel
}

func GetStatsModel() models.StatsModel {
	sModel := models.StatsModel{Db: database.Db}

	return sModel
}

func InitializeUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	vModel := GetVideoModel()

	video, err := vModel.CreateVideo(models.Video{
		UploadComplete: false,
	})

	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}

	utils.Success(w, http.StatusCreated, map[string]interface{}{
		"message": "Video Upload Initialized",
		"video":   video,
	})

}

func UploadChunk(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	videoID := p.ByName("videoID")

	videoModel := GetVideoModel()
	_, err := videoModel.FindVideoByID(videoID)
	if err != nil {
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	fiftyMB := int64(50 << 20) // Max Size of 50MB
	err = r.ParseMultipartForm(fiftyMB)
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}

	chunk, _, err := r.FormFile("chunk")
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	chunkNumberStr := r.FormValue("chunk_number")
	if chunkNumberStr == "" {
		utils.Fail(w, http.StatusUnprocessableEntity, map[string]interface{}{"chunk_number": "Chunk number is required"})
		return
	}

	defer func(chunk multipart.File) {
		err := chunk.Close()
		if err != nil {
			log.Println("Unable to close file: ", err)
		}
	}(chunk)

	err = ensureRootDirExists(videoID)
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	uploadPath := filepath.Join(getVideoRootDir(videoID), chunkNumberStr)

	f, err := os.OpenFile(uploadPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, chunk)
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	sModel := GetStatsModel()

	stats, err := sModel.NewChunkUpload(videoID)

	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}

	utils.Success(w, http.StatusOK, map[string]interface{}{"message": "Uploading", "stats": stats})
}

func CompleteUpload(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	videoID := p.ByName("videoID")

	videoModel := GetVideoModel()
	_, err := videoModel.FindVideoByID(videoID)
	if err != nil {
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	statsModel := GetStatsModel()
	stats, err := statsModel.FindStatsForVideo(videoID)
	if err != nil {
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}
	var buf bytes.Buffer

	rootDir := getVideoRootDir(videoID)
	for i := 1; i <= stats.NumChunks; i++ {
		fileBytes, err := os.ReadFile(filepath.Join(rootDir, fmt.Sprintf("%v", i)))

		cType := http.DetectContentType(fileBytes)
		log.Println(cType)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				//	Chunk does not exist. Either the chunk wasn't uploaded or wasn't saved correctly. We ignore.
				continue
			} else {
				log.Println("err")
				utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
				return
			}

		}

		// Store it in memory.
		buf.Write(fileBytes)

		// Limiting the usage to 500 MB of RAM.
		fiveHundredMB := 1024 * 1024 * 500
		if buf.Len() >= fiveHundredMB {
			log.Println("breaking due to very large video:", buf.Len())
			break
		}
	}

	finalPath := filepath.Join(rootDir, "final")
	f, err := os.OpenFile(finalPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, &buf)
	if err != nil {
		log.Println("unable to store final copy")
		utils.Fail(w, http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		return
	}

	video, err := videoModel.CompleteUpload(videoID)
	if err != nil {
		log.Println("unable to update db")
		utils.Fail(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}

	log.Println("Starting transcription in the background.")
	utils.Success(w, http.StatusOK, map[string]interface{}{"message": "Upload complete. Transcription in progress", "video": video})
}

func ensureRootDirExists(videoID string) error {
	rootDir := getVideoRootDir(videoID)
	stat, err := os.Stat(rootDir)
	log.Println(stat)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		//	Directory doesn't exist
		//	We create it.
		// Permission is just read and write
		return os.Mkdir(rootDir, 0777)
	}

	return err
}

func getVideoRootDir(videoID string) string {
	return filepath.Join(config.Environment.UploadPath, videoID)
}
