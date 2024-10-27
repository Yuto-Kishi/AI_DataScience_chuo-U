from gpiozero import Button, OutputDevice, Servo
from RPLCD.i2c import CharLCD
import face_recognition
import cv2
import random
import sqlite3
from datetime import datetime
from time import sleep

# SQLiteデータベースの設定
conn = sqlite3.connect('../user_registration.db')  # パスを正しく設定
cursor = conn.cursor()

# テーブルが存在しない場合の作成
cursor.execute('''CREATE TABLE IF NOT EXISTS study_data (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    date TEXT,
                    detection_time REAL,
                    study_time REAL,
                    focus_score REAL
                )''')

# GPIOピンの設定
ROWS = [5, 6, 13, 19]  # 行ピン
COLS = [26, 21, 20, 16]  # 列ピン
SERVO_PIN = 17
SWITCH_PIN = 18

# キーパッドマッピング
keys = [
    ['1', '2', '3', '3'],
    ['4', '5', '6', '6'],
    ['7', '8', '9', '9'],
    ['*', '0', '#', '#']
]

# デバイスの設定
row_buttons = [Button(pin, pull_up=False) for pin in ROWS]
col_devices = [OutputDevice(pin) for pin in COLS]
servo = Servo(SERVO_PIN)
lcd = CharLCD('PCF8574', 0x27)

# 顔検出変数
video_capture = cv2.VideoCapture(0)
face_detection_time = 0
face_present = False
detection_start_time = None

# ランダム4桁の番号生成関数
def generate_target_code():
    return f"{random.randint(1000, 9999)}"

# データ保存関数
def record_study_data(detection_time, study_time, focus_score):
    date = datetime.now().strftime("%Y-%m-%d")
    cursor.execute("INSERT INTO study_data (date, detection_time, study_time, focus_score) VALUES (?, ?, ?, ?)",
                   (date, detection_time, study_time, focus_score))
    conn.commit()

# ユーザーID入力
def enter_user_id():
    lcd.clear()
    lcd.write_string("User ID: ")
    user_id_digits = []

    while True:
        for col_idx, col in enumerate(COLS):
            col_devices[col_idx].on()
            for row_idx, row_button in enumerate(row_buttons):
                if row_button.is_pressed:
                    sleep(0.1)
                    if row_button.is_pressed:
                        digit = keys[row_idx][col_idx]
                        if digit == '#':  # "#"で入力終了
                            user_id = ''.join(user_id_digits)
                            print(f"User ID: {user_id}")
                            return user_id
                        elif len(user_id_digits) < 4:
                            user_id_digits.append(digit)
                            lcd.cursor_pos = (0, 9)
                            lcd.write_string(''.join(user_id_digits).ljust(4))
                    sleep(0.5)
            col_devices[col_idx].off()
            sleep(0.05)

# カウントダウンと顔認識
def start_timer_and_face_recognition():
    global face_detection_time, face_present, detection_start_time
    servo.mid()
    lcd.clear()
    lcd.write_string("Countdown...")

    for seconds_left in range(15, 0, -1):
        lcd.cursor_pos = (0, 0)
        lcd.write_string(f"Time left: {seconds_left}s   ")
        sleep(1)

        ret, frame = video_capture.read()
        if not ret:
            print("カメラから映像を取得できませんでした。")
            break

        rgb_frame = frame[:, :, ::-1]
        face_locations = face_recognition.face_locations(rgb_frame, model="hog")

        if face_locations:
            if not face_present:
                detection_start_time = datetime.now()
                face_present = True
        else:
            if face_present:
                if detection_start_time:
                    detection_end_time = datetime.now()
                    face_detection_time += (detection_end_time - detection_start_time).total_seconds()
                face_present = False

        for (top, right, bottom, left) in face_locations:
            cv2.rectangle(frame, (left, top), (right, bottom), (0, 255, 0), 2)

        cv2.imshow('Camera', frame)
        if cv2.waitKey(1) & 0xFF == ord('q'):
            break

    if face_present and detection_start_time:
        detection_end_time = datetime.now()
        face_detection_time += (detection_end_time - detection_start_time).total_seconds()

# キーパッド入力確認
def keypad_input_check(target_code):
    input_digits = []
    lcd.clear()
    lcd.write_string(f"Code: {target_code}")
    lcd.cursor_pos = (1, 0)
    lcd.write_string("Enter: ")

    while True:
        for col_idx, col in enumerate(COLS):
            col_devices[col_idx].on()
            for row_idx, row_button in enumerate(row_buttons):
                if row_button.is_pressed:
                    sleep(0.1)
                    if row_button.is_pressed:
                        digit = keys[row_idx][col_idx]
                        input_digits.append(digit)
                        lcd.cursor_pos = (1, 7)
                        lcd.write_string(''.join(input_digits).ljust(4))

                        if len(input_digits) == 4:
                            entered_code = ''.join(input_digits)
                            if entered_code == target_code:
                                lcd.clear()
                                lcd.write_string("Success!")
                                servo.min()
                                focus_score = (face_detection_time / 15) * 100
                                record_study_data(face_detection_time, 15, focus_score)
                                return
                            else:
                                lcd.clear()
                                lcd.write_string(f"Code: {target_code}")
                                lcd.cursor_pos = (1, 0)
                                lcd.write_string("Enter: ")
                            input_digits = []
                    sleep(0.5)
            col_devices[col_idx].off()
            sleep(0.05)

# メイン処理
try:
    user_id = enter_user_id()
    start_timer_and_face_recognition()
    cv2.destroyAllWindows()

    target_code = generate_target_code()
    lcd.clear()
    lcd.write_string(f"Code: {target_code}")
    lcd.cursor_pos = (1, 0)
    lcd.write_string("Enter: ")

    keypad_input_check(target_code)

except KeyboardInterrupt:
    print("プログラムが終了されました。")
finally:
    video_capture.release()
    cv2.destroyAllWindows()
    conn.close()

