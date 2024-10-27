package main

import (
    "net/http"
    "monitering_study/models"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "time"
)

var db *gorm.DB

// 学習記録モデル
type StudyRecord struct {
    ID              uint `gorm:"primaryKey"`
    UserID          uint
    StartTime       time.Time
    EndTime         time.Time
    FocusedDuration time.Duration
}

func main() {
    var err error
    db, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&models.User{}, &StudyRecord{}) // 学習記録モデルのマイグレーション

    r := gin.Default()

    r.LoadHTMLGlob("templates/*")

    r.GET("/register", showRegisterPage)
    r.POST("/register", registerUser)
    r.GET("/login", showLoginPage)
    r.POST("/login", loginUser)
    r.GET("/dashboard", showDashboardPage) // ダッシュボードのルートを追加
    r.POST("/logout", logoutUser) // ログアウトのルートを追加

    r.Run(":8080")
}

func showRegisterPage(c *gin.Context) {
    c.HTML(http.StatusOK, "register.html", nil)
}

func registerUser(c *gin.Context) {
    email := c.PostForm("email")
    username := c.PostForm("username")

    var existingUser models.User
    db.Where("email = ?", email).First(&existingUser)

    if existingUser.Email != "" {
        c.String(http.StatusBadRequest, "すでに登録されています。ログインしてください！")
        return
    }

    user := models.User{Email: email, Username: username}
    result := db.Create(&user)

    if result.Error != nil {
        c.String(http.StatusBadRequest, "ユーザー登録に失敗しました。")
        return
    }

    c.Redirect(http.StatusSeeOther, "/dashboard") // ダッシュボードにリダイレクト
}

func showLoginPage(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", nil)
}

func loginUser(c *gin.Context) {
    email := c.PostForm("email")
    username := c.PostForm("username")

    var user models.User
    result := db.Where("email = ? AND username = ?", email, username).First(&user)

    if result.Error != nil {
        c.String(http.StatusBadRequest, "ユーザーが見つかりません。登録が必要です。")
        return
    }

    c.Redirect(http.StatusSeeOther, "/dashboard") // ダッシュボードにリダイレクト
}

func showDashboardPage(c *gin.Context) {
    // ダッシュボードに必要なデータを取得
    var records []StudyRecord
    db.Find(&records)

    // 例として、現在の日付と時刻を取得
    currentDate := time.Now().Format("2006-01-02")
    c.HTML(http.StatusOK, "dashboard.html", gin.H{
        "currentDate": currentDate,
        "records":     records,
    })
}

func logoutUser(c *gin.Context) {
    // ログアウトの処理を行う（セッション管理をしている場合）
    c.Redirect(http.StatusSeeOther, "/login") // ログイン画面にリダイレクト
}

