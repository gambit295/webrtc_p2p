package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "sync"
    "time"
    "github.com/gorilla/websocket"
)

type Room struct {
    Teacher *Peer `json:"teacher"`
    Student *Peer `json:"student"`
}

type Peer struct {
    ID       string `json:"id"`
    Role     string `json:"role"`
    Conn     *websocket.Conn
}

type SignalMessage struct {
    Type    string          `json:"type"`
    Role    string          `json:"role,omitempty"`
    Target  string          `json:"target,omitempty"`
    SDP     string          `json:"sdp,omitempty"`
    ICE     json.RawMessage `json:"ice,omitempty"`
    Message string          `json:"message,omitempty"`
    File    *FileInfo       `json:"file,omitempty"`
    RoomID  string          `json:"roomId,omitempty"`
}

type FileInfo struct {
    Name string `json:"name"`
    Size int64  `json:"size"`
    Type string `json:"type"`
    Data []byte `json:"data,omitempty"`
}

var (
    rooms = make(map[string]*Room)
    mutex = &sync.RWMutex{}
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
        // Увеличиваем размер буфера для файлов
        ReadBufferSize:  1024 * 1024, // 1MB
        WriteBufferSize: 1024 * 1024, // 1MB
    }
)

// Настройка логирования в файл для Docker
func init() {
    // Создаем директорию для логов если её нет
    if err := os.MkdirAll("./logs", 0755); err == nil {
        // Открываем файл лога
        logFile, err := os.OpenFile("./logs/p2p-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err == nil {
            log.SetOutput(logFile)
        }
    }
    
    // Добавляем временную метку
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
    log.Println("=== P2P Learning Server Starting ===")
    log.Printf("Go Version: %s", os.Getenv("GOLANG_VERSION"))
    
    // Запускаем HTTP редирект сервер
    go startHTTPServer()
    
    // Основной HTTPS сервер
    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir("./static")))
    mux.HandleFunc("/ws", handleWebSocket)
    mux.HandleFunc("/health", healthCheck)
    mux.HandleFunc("/metrics", metricsHandler)

    httpsServer := &http.Server{
        Addr:         ":443",
        Handler:      mux,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Проверяем наличие сертификатов
    if _, err := os.Stat("./certs/cert.pem"); os.IsNotExist(err) {
        log.Fatal("SSL certificate not found. Please generate certificates first.")
    }
    if _, err := os.Stat("./certs/key.pem"); os.IsNotExist(err) {
        log.Fatal("SSL key not found. Please generate certificates first.")
    }

    log.Println("HTTPS server starting on :443")
    log.Println("HTTP server on :80 will redirect to HTTPS")
    log.Println("WebSocket endpoint: wss://<server>/ws")
    log.Println("Health check: https://<server>/health")
    
    // Запускаем HTTPS сервер
    if err := httpsServer.ListenAndServeTLS("./certs/cert.pem", "./certs/key.pem"); err != nil {
        log.Fatal("HTTPS server failed:", err)
    }
}

func startHTTPServer() {
    httpRedirect := &http.Server{
        Addr: ":80",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            target := "https://" + r.Host + r.URL.Path
            if r.URL.RawQuery != "" {
                target += "?" + r.URL.RawQuery
            }
            log.Printf("Redirect: %s -> %s", r.URL.Path, target)
            http.Redirect(w, r, target, http.StatusMovedPermanently)
        }),
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }
    
    log.Println("HTTP redirect server started on :80")
    if err := httpRedirect.ListenAndServe(); err != nil {
        log.Printf("HTTP server stopped: %v", err)
    }
}

// Health check endpoint для Docker
func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "time":   time.Now().Unix(),
        "rooms":  len(rooms),
    })
}

// Metrics endpoint (упрощенный)
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    mutex.RLock()
    defer mutex.RUnlock()
    
    activeRooms := 0
    activeTeachers := 0
    activeStudents := 0
    
    for _, room := range rooms {
        activeRooms++
        if room.Teacher != nil {
            activeTeachers++
        }
        if room.Student != nil {
            activeStudents++
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "active_rooms":    activeRooms,
        "active_teachers": activeTeachers,
        "active_students": activeStudents,
        "total_peers":     activeTeachers + activeStudents,
    })
}

// Остальной код без изменений (handleWebSocket, notifyPeerJoined, forwardSignal, forwardToPeer, removePeer)
// ... (весь остальной код из предыдущего main.go)
