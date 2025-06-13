package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os" // os 패키지를 임포트하여 환경 변수를 읽습니다.
	"sync"
	"time"
)

// ... (LogEntry 구조체, logBuffer, mu 등 기존 코드 동일) ...
type LogEntry struct {
    Timestamp  string `json:"timestamp"`
    File       string `json:"file"`
    ThreadID   string `json:"threadId"`
    Protocol   string `json:"protocol"`
    BytesSent  int    `json:"bytesSent"`
    TargetIP   string `json:"targetIp"`
    PacketInfo string `json:"packetInfo"`
}

var (
    logBuffer []LogEntry
    mu        sync.Mutex
)

func logPacketSent(tid int, bytes int, ip string, packet string) {
    formattedLog := fmt.Sprintf("%s main.go:211: [TID:%d] [UDP] Sent %d bytes to %s (%s)",
        time.Now().Format("2006/01/02 15:04:05"), tid, bytes, ip, packet)
    log.Println(formattedLog)

    mu.Lock()
    defer mu.Unlock()

    logBuffer = append(logBuffer, LogEntry{
        Timestamp:  time.Now().Format("2006/01/02 15:04:05"),
        File:       "main.go:211",
        ThreadID:   fmt.Sprintf("%d", tid),
        Protocol:   "UDP",
        BytesSent:  bytes,
        TargetIP:   ip,
        PacketInfo: packet,
    })
    if len(logBuffer) > 200 {
        logBuffer = logBuffer[len(logBuffer)-200:]
    }
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("index.html")
    if err != nil {
        log.Printf("템플릿 파싱 에러: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}

func apiLogsHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    currentLogs := make([]LogEntry, len(logBuffer))
    copy(currentLogs, logBuffer)
    mu.Unlock()

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(currentLogs); err != nil {
        log.Printf("JSON 인코딩 에러: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func main() {
	go func() {
		for i := 0; ; i++ {
			time.Sleep(500 * time.Millisecond)
			tid := (i % 1000) + 100
			packetInfo := fmt.Sprintf("Packet %d/%d", (i % 85481) + 1, 85481)
			logPacketSent(tid, 1472, "2406:da14:540:e901::6e:4:443", packetInfo)
		}
	}()

	http.HandleFunc("/", logsHandler)
	http.HandleFunc("/api/logs", apiLogsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Render 환경 변수에서 PORT를 가져오거나, 기본값으로 8080을 사용합니다.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 로컬 개발용 기본 포트
	}

	fmt.Printf("웹 서버가 0.0.0.0:%s 에서 실행 중입니다.\n", port)
	// 모든 네트워크 인터페이스에서 수신 대기하도록 ":PORT"를 사용합니다.
	log.Fatal(http.ListenAndServe(":"+port, nil))
}