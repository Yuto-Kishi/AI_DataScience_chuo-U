package models

import (
    "gorm.io/gorm"
)

// User はユーザーモデル
type User struct {
    gorm.Model
    Email    string `gorm:"unique"`
    Username string
}

// StudyRecord は学習記録モデル
type StudyRecord struct {
    gorm.Model
    UserID          uint
    StartTime       string
    EndTime         string
    ConcentrationTime int
}

