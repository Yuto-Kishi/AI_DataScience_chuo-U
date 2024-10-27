package main

import (
    "database/sql"
    "html/template"
    "log"
    "net/http"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type DashboardData struct {
    CurrentDate string
    StudyData   []map[string]interface{}
}

func init() {
    var err error
    db, err = sql.Open("sqlite3", "./user_registration.db")
    if err != nil {
        log.Fatal(err)
    }

    createUserTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE,
        username TEXT NOT NULL UNIQUE
    );`
    createStudyTable := `
    CREATE TABLE IF NOT EXISTS study_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT,
        detection_time REAL,
        study_time REAL,
        focus_score REAL
    );`

    _, err = db.Exec(createUserTable)
    if err != nil {
        log.Fatal(err)
    }
    _, err = db.Exec(createStudyTable)
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    http.HandleFunc("/", registrationPage)
    http.HandleFunc("/register", registerUser)
    http.HandleFunc("/login", loginPage)
    http.HandleFunc("/authenticate", authenticateUser)
    http.HandleFunc("/dashboard", dashboardPage)
    http.HandleFunc("/logout", logout)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// ユーザー登録ページのハンドラー
func registrationPage(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/register.html"))
    tmpl.Execute(w, nil)
}

// ユーザーをデータベースに登録するハンドラー
func registerUser(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        email := r.FormValue("email")
        username := r.FormValue("username")

        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", email).Scan(&exists)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        if exists {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        _, err = db.Exec("INSERT INTO users (email, username) VALUES (?, ?)", email, username)
        if err != nil {
            http.Error(w, "Failed to register user", http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/login", http.StatusSeeOther)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

// ログインページのハンドラー
func loginPage(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/login.html"))
    tmpl.Execute(w, nil)
}

// ログイン認証のハンドラー
func authenticateUser(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        email := r.FormValue("email")
        username := r.FormValue("username")

        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=? AND username=?)", email, username).Scan(&exists)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        if exists {
            http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
        } else {
            http.Redirect(w, r, "/", http.StatusSeeOther)
        }
    } else {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
    }
}

// ダッシュボードページのハンドラー
func dashboardPage(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))

    currentDate := time.Now().Format("06/01/02") // 今日の日付 (YY/MM/DD 形式)

    rows, err := db.Query("SELECT date, detection_time, study_time, focus_score FROM study_data ORDER BY date DESC")
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var data []map[string]interface{}
    for rows.Next() {
        var date string
        var detectionTime, studyTime, focusScore float64
        rows.Scan(&date, &detectionTime, &studyTime, &focusScore)
        record := map[string]interface{}{
            "Date":          date,
            "DetectionTime": detectionTime,
            "StudyTime":     studyTime,
            "FocusScore":    focusScore,
        }
        data = append(data, record)
    }

    tmpl.Execute(w, DashboardData{CurrentDate: currentDate, StudyData: data})
}

// ログアウトのハンドラー
func logout(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

