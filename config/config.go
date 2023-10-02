package config

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	DbUser     string `json:"DB_USER"`
	DbPassword string `json:"DB_PASSWORD"`
	DbHost     string `json:"DB_HOST"`
	DbName     string `json:"DB_NAME"`
	DbPort     string `json:"DB_PORT"`

	Port       string `json:"APP_PORT"`
	UploadPath string `json:"UPLOAD_PATH"`
}

var Environment Config
var loaded = false

// Load Should be called once.
func Load() {
	data, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("Unable to read environment variables:", err)
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Unable to parse environment variables:", err)
	}

	err = json.Unmarshal(marshalled, &Environment)

	if err != nil {
		log.Fatal("Unable to parse environment variables:", err)
	}

	loaded = true
	log.Println("Config Loaded Successfully")
}

func IsLoaded() bool {
	return loaded
}
