package models

import "gorm.io/gorm"

type LearningRecord struct {
    gorm.Model
    UserID          uint   `gorm:"index"` // ユーザーIDを追加
    Date            string // 学習日
    StartTime       string // 開始時間
    EndTime         string // 終了時間
    FocusedDuration int    // 集中時間（分単位）
}

