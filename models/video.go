package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Video struct {
	Id             string `json:"id" gorm:"type:varchar(255);primaryKey"`
	UploadComplete bool   `json:"upload_complete" `
}

type VideoModel struct {
	Db *gorm.DB
}

func (e *Video) BeforeCreate(tx *gorm.DB) error {
	e.Id = uuid.NewString()

	return nil
}

func (vM *VideoModel) CreateVideo(video Video) (*Video, error) {
	result := vM.Db.Create(&video)

	if result.Error != nil {
		return nil, result.Error
	}

	return &video, nil
}

func (vM *VideoModel) FindVideoByID(videoID string) (*Video, error) {
	var video Video

	err := vM.Db.Where(&Video{Id: videoID}).First(&video).Error

	if err != nil {
		return nil, err
	}

	return &video, nil
}

func (vM *VideoModel) CompleteUpload(videoID string) (*Video, error) {
	err := vM.Db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		err := tx.Where(&Video{Id: videoID}).Updates(&Video{UploadComplete: true}).Error

		if err != nil {
			return err
		}

		return tx.Where(&UploadStatistic{VideoID: videoID}).Updates(&UploadStatistic{UploadCompletedAt: &now}).Error
	})

	if err != nil {
		return nil, err
	}

	return vM.FindVideoByID(videoID)
}
