package main

import (
    "database/sql"
    "html/template"
    "net/http"
    "time"
    _ "github.com/mattn/go-sqlite3"
)

type StudyData struct {
    Date          string
    DetectionTime float64
    StudyTime     float64
    FocusScore    float64
}

var dashboardTemplate = template.Must(template.ParseFiles("templates/dashboard.html"))

func dashboardPage(w http.ResponseWriter, r *http.Request) {
    db, _ := sql.Open("sqlite3", "./user_registration.db")
    defer db.Close()

    rows, _ := db.Query("SELECT date, detection_time, study_time, focus_score FROM study_data ORDER BY date DESC")
    var data []StudyData
    for rows.Next() {
        var record StudyData
        rows.Scan(&record.Date, &record.DetectionTime, &record.StudyTime, &record.FocusScore)
        data = append(data, record)
    }

    dashboardTemplate.Execute(w, data)
}

func logout(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

