package models

import (
	"gorm.io/gorm"
	"time"
)

type StatsModel struct {
	Db *gorm.DB
}

type UploadStatistic struct {
	gorm.Model

	VideoID           string     `json:"video_id" gorm:"type:varchar(255)"`
	NumChunks         int        `json:"num_chunks"`
	LastChunkAddedAt  *time.Time `json:"last_chunk_added_at"`
	UploadCompletedAt *time.Time `json:"upload_completed_at"`

	Video Video `json:"video,omitempty"`
}

func (sM *StatsModel) FindStatsForVideo(videoID string) (*UploadStatistic, error) {
	var stats UploadStatistic

	result := sM.Db.FirstOrCreate(&stats, &UploadStatistic{VideoID: videoID})

	if result.Error != nil {
		return nil, result.Error
	}

	return &stats, nil
}

func (sM *StatsModel) NewChunkUpload(videoID string) (*UploadStatistic, error) {
	stats, err := sM.FindStatsForVideo(videoID)
	if err != nil {
		return nil, err

	}

	now := time.Now()

	err = sM.Db.Where(&UploadStatistic{VideoID: videoID}).Updates(&UploadStatistic{
		NumChunks:        stats.NumChunks + 1,
		LastChunkAddedAt: &now,
	}).Error

	if err != nil {
		return nil, err
	}

	return sM.FindStatsForVideo(videoID)
}

func (sM *StatsModel) CompleteUpload(videoID string) (*UploadStatistic, error) {
	now := time.Now()
	err := sM.Db.Where(&UploadStatistic{VideoID: videoID}).Updates(&UploadStatistic{
		UploadCompletedAt: &now,
	}).Error

	if err != nil {
		return nil, err
	}

	return sM.FindStatsForVideo(videoID)
}
