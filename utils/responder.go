package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func Success(w http.ResponseWriter, statusCode int, data map[string]interface{}) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		Fail(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	if statusCode != http.StatusNoContent {
		_, err = w.Write(jsonResponse)
	}
	if err != nil {
		log.Println("unhandled error:", err)
	}
}

func Fail(w http.ResponseWriter, statusCode int, data map[string]interface{}) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		log.Println("unhandled error:", err)
		return
	}

	log.Println("Handling failed response:", string(jsonResponse))

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println("unhandled error:", err)
	}

}
