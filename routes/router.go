package routes

import (
	"github.com/julienschmidt/httprouter"
	"jameesjohn.com/uploader-api/controllers"
	"net/http"
)

func Router() http.Handler {
	router := httprouter.New()

	router.GET("/watch/:videoID", controllers.WatchVideo)
	router.GET("/upload/init", controllers.InitializeUpload)
	router.POST("/upload/chunk/:videoID", controllers.UploadChunk)
	router.POST("/upload/complete/:videoID", controllers.CompleteUpload)

	return router
}
