package main

import (
	"embed"
	"time"

	"github.com/gorilla/websocket"
)

const serverAddress = ":8080"

const (
	BoardSize         = 15
	TotalCells        = BoardSize * BoardSize
	MaxNicknameLength = 10
	TimeoutDuration   = 60 * time.Second
	WinningCount      = 5
)

const (
	black   uint8 = 1
	white   uint8 = 2
	emptied uint8 = 0
)

const (
	StatusUser1Timeout = 3
	StatusUser2Timeout = 2
	StatusErrorReading = 4
)

const (
	WebSocketPingType = "ping"
	WebSocketPongType = "pong"
)

const (
	HTMLPath   = "HTML"
	ConfigPath = "CONFIGS"
	ImagePath  = "IMAGE"
	SoundPath  = "SOUND"
)

//go:embed CONFIGS
var CONFIGS embed.FS

//go:embed HTML
var HTML embed.FS

//go:embed IMAGE
var IMAGE embed.FS

//go:embed SOUND
var SOUND embed.FS

type OmokRoom struct {
	board_15x15 [TotalCells]uint8
	user1       user
	user2       user
	spectators  []*websocket.Conn
}

type user struct {
	ws       *websocket.Conn
	check    bool
	nickname string
}

type Message struct {
	Data      interface{} `json:"data,omitempty"`
	YourColor interface{} `json:"YourColor,omitempty"`
	Message   interface{} `json:"message,omitempty"`
	NumUsers  interface{} `json:"numUsers,omitempty"`
	Nickname  interface{} `json:"nickname,omitempty"`
}

type SpectatorMessage struct {
	Board interface{} `json:"board,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Color interface{} `json:"color,omitempty"`
	User1 interface{} `json:"user1,omitempty"`
	User2 interface{} `json:"user2,omitempty"`
}

var (
	upgrader         = websocket.Upgrader{}
	rooms            []*OmokRoom
	sockets          []*websocket.Conn
	connectionsCount = 0
)
