package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
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
    Type    string          `json:"type"` // offer, answer, candidate, role, disconnect, chat, file
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
    }
)

func main() {
    // Статические файлы клиента
    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    // WebSocket для сигналинга (WSS)
    http.HandleFunc("/ws", handleWebSocket)
    
    log.Println("Signal server starting on https://localhost:8443")
    log.Println("Используйте https://<IP-адрес-сервера>:8443 для доступа с других устройств")
    
    // Запускаем HTTPS сервер
    err := http.ListenAndServeTLS(":8443", "certs/cert.pem", "certs/key.pem", nil)
    if err != nil {
        log.Fatal("ListenAndServeTLS: ", err)
    }
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("Upgrade failed:", err)
        return
    }
    defer conn.Close()

    var currentPeer *Peer
    var currentRoomID string
    
    for {
        var msg SignalMessage
        err := conn.ReadJSON(&msg)
        if err != nil {
            log.Println("Read error:", err)
            // Удаляем пира при отключении
            if currentPeer != nil {
                removePeer(currentRoomID, currentPeer)
            }
            break
        }

        switch msg.Type {
        case "join":
            // Регистрация в комнате
            roomID := "room1" // Для простоты используем одну комнату
            currentRoomID = roomID
            
            currentPeer = &Peer{
                ID:   conn.RemoteAddr().String(),
                Role: msg.Role,
                Conn: conn,
            }
            
            mutex.Lock()
            if _, exists := rooms[roomID]; !exists {
                rooms[roomID] = &Room{}
            }
            
            if msg.Role == "teacher" {
                // Если учитель уже есть, заменяем его
                if rooms[roomID].Teacher != nil {
                    oldTeacher := rooms[roomID].Teacher
                    oldTeacher.Conn.WriteJSON(SignalMessage{
                        Type: "peer_left",
                    })
                }
                rooms[roomID].Teacher = currentPeer
                log.Println("Teacher joined:", currentPeer.ID)
            } else {
                // Если ученик уже есть, заменяем его
                if rooms[roomID].Student != nil {
                    oldStudent := rooms[roomID].Student
                    oldStudent.Conn.WriteJSON(SignalMessage{
                        Type: "peer_left",
                    })
                }
                rooms[roomID].Student = currentPeer
                log.Println("Student joined:", currentPeer.ID)
            }
            mutex.Unlock()
            
            // Уведомляем о подключении
            notifyPeerJoined(roomID, msg.Role)
            
        case "offer", "answer", "candidate":
            // Пересылка сигналов между пирами
            if currentRoomID != "" {
                forwardSignal(currentRoomID, msg)
            }
            
        case "chat":
            // Пересылка сообщений чата
            if currentRoomID != "" {
                forwardToPeer(currentRoomID, msg.Target, msg)
            }
            
        case "file":
            // Пересылка файлов (в реальном проекте лучше через P2P)
            if currentRoomID != "" {
                forwardToPeer(currentRoomID, msg.Target, msg)
            }
            
        case "disconnect":
            if currentPeer != nil && currentRoomID != "" {
                removePeer(currentRoomID, currentPeer)
            }
            return
        }
    }
}

func notifyPeerJoined(roomID, role string) {
    mutex.RLock()
    room := rooms[roomID]
    mutex.RUnlock()
    
    if room == nil {
        return
    }
    
    // Уведомляем учителя о ученике
    if role == "student" && room.Teacher != nil {
        err := room.Teacher.Conn.WriteJSON(SignalMessage{
            Type: "peer_joined",
            Role: "student",
        })
        if err != nil {
            log.Println("Error notifying teacher:", err)
        }
    }
    
    // Уведомляем ученика об учителе
    if role == "teacher" && room.Student != nil {
        err := room.Student.Conn.WriteJSON(SignalMessage{
            Type: "peer_joined",
            Role: "teacher",
        })
        if err != nil {
            log.Println("Error notifying student:", err)
        }
    }
}

func forwardSignal(roomID string, msg SignalMessage) {
    mutex.RLock()
    room := rooms[roomID]
    mutex.RUnlock()
    
    if room == nil {
        return
    }
    
    var targetPeer *Peer
    if msg.Target == "student" {
        targetPeer = room.Student
    } else if msg.Target == "teacher" {
        targetPeer = room.Teacher
    }
    
    if targetPeer != nil {
        err := targetPeer.Conn.WriteJSON(msg)
        if err != nil {
            log.Println("Error forwarding signal:", err)
        }
    }
}

func forwardToPeer(roomID, targetRole string, msg SignalMessage) {
    mutex.RLock()
    room := rooms[roomID]
    mutex.RUnlock()
    
    if room == nil {
        return
    }
    
    var targetPeer *Peer
    if targetRole == "teacher" {
        targetPeer = room.Teacher
    } else {
        targetPeer = room.Student
    }
    
    if targetPeer != nil {
        err := targetPeer.Conn.WriteJSON(msg)
        if err != nil {
            log.Println("Error forwarding message:", err)
        }
    }
}

func removePeer(roomID string, peer *Peer) {
    mutex.Lock()
    defer mutex.Unlock()
    
    room := rooms[roomID]
    if room == nil {
        return
    }
    
    if room.Teacher != nil && room.Teacher.ID == peer.ID {
        room.Teacher = nil
        if room.Student != nil {
            err := room.Student.Conn.WriteJSON(SignalMessage{
                Type: "peer_left",
            })
            if err != nil {
                log.Println("Error notifying student about teacher left:", err)
            }
        }
        log.Println("Teacher removed:", peer.ID)
    } else if room.Student != nil && room.Student.ID == peer.ID {
        room.Student = nil
        if room.Teacher != nil {
            err := room.Teacher.Conn.WriteJSON(SignalMessage{
                Type: "peer_left",
            })
            if err != nil {
                log.Println("Error notifying teacher about student left:", err)
            }
        }
        log.Println("Student removed:", peer.ID)
    }
    
    // Удаляем пустую комнату
    if room.Teacher == nil && room.Student == nil {
        delete(rooms, roomID)
        log.Println("Room removed:", roomID)
    }
}
