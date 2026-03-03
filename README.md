# P2P Learning with Pion WebRTC 🎓

[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8.svg)](https://golang.org/)
[![WebRTC](https://img.shields.io/badge/WebRTC-Pion-2E8B57.svg)](https://github.com/pion/webrtc)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**P2P Learning** is a web application for direct peer-to-peer interaction between teacher and student, built on Pion WebRTC and Golang. The application enables video conferencing with chat, file sharing, and exclusive features for teachers.

## ✨ Features

### 🎥 Video Communication
- Direct P2P connection (WebRTC)
- Large remote video, compact local video
- Camera and microphone status indicators for both participants
- Screen sharing for teachers

### 💬 Chat
- Collapsible chat panel (vertical strip when minimized)
- Real-time text messaging
- File sharing
- Automatic notifications when devices are turned on/off

### 👨‍🏫 Teacher Features
- **"Attention!"** button — when pressed, the student sees:
  - Chat automatically expands (if collapsed)
  - A prominent ⚠️ ATTENTION ⚠️ message appears
- Control over own camera, microphone, and screen sharing
- Teacher stays in session even after student disconnects

### 🧑‍🎓 Student Features
- Control over own camera and microphone
- Teacher waiting indicator (shown only when teacher is not connected)
- Automatic connection when teacher appears

### 🛠 Technical Highlights
- **Smart Waiting**: indicator only appears if teacher is truly offline
- **Reliable Reconnection**: automatic reconnection on connection loss
- **Status Bar**: displays WebSocket, P2P, and media device states
- **Debug Panel**: hidden under a spoiler with detailed application logs

## 📋 Requirements

- Go 1.16 or higher
- Modern browser with WebRTC support (Chrome, Firefox, Safari, Edge)
- Access to camera and microphone
- For HTTPS — self-signed or valid SSL certificates

## 🚀 Installation & Running

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/p2p-learning.git
cd p2p-learning
```

2. **Create SSL certificates** (for HTTPS)
```bash
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes
```

3. **Run the server**
```bash
go run main.go
```

4. **Open the application**
```
https://localhost:8443
```

## 🏗 Project Structure

```
p2p-learning/
├── main.go              # Golang signaling server (WebSocket)
├── static/              # Static files (HTML, CSS, JS)
│   └── index.html       # Client application
├── certs/               # SSL certificates
│   ├── cert.pem
│   └── key.pem
└── README.md            # Documentation
```

## 🔧 Configuration

### Server (main.go)
- **Port**: 8443 (HTTPS)
- **WebSocket endpoint**: `/ws`
- **STUN servers**: Google Public STUN (can be changed in `ICE_SERVERS`)

### Client (index.html)
- Automatic protocol detection (wss/ws)
- Responsive design
- Support for all modern browsers

## 📱 Screenshots

*(Add your screenshots here)*

## 🧪 Testing

To test, open the application in two different browsers or on two different devices:
1. In one browser, click **"I am Teacher"**
2. In the other browser, click **"I am Student"**
3. Allow camera and microphone access
4. Enjoy your session!

## 🔐 Security

- All connections use HTTPS/WSS
- P2P traffic is encrypted (WebRTC DTLS/SRTP)
- Signaling server doesn't store any data
- Certificates can be replaced with valid ones for production

## 🛣 Roadmap

- [x] Basic video conferencing
- [x] Chat with file sharing
- [x] "Attention" button
- [x] Smart teacher waiting indicator
- [ ] Session recording
- [ ] Virtual whiteboard
- [ ] Password-protected rooms
- [ ] Mobile app

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` file for details.

## 🙏 Acknowledgments

- [Pion WebRTC](https://github.com/pion/webrtc) — excellent Go WebRTC implementation
- [Gorilla WebSocket](https://github.com/gorilla/websocket) — reliable WebSocket library
- All contributors and users!

---

**P2P Learning** — made with ❤️ for convenient online education
