package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed HTML/1.html
var html1 string

//go:embed IMAGE/favicon.ico
var favicon string

//go:embed SOUND
var SOUND embed.FS

const max = 100

const black uint8 = 1
const white uint8 = 2
const emptied uint8 = 0

type OmokRoom struct {
	board_15x15 [225]uint8
	uesr_1      user
	uesr_2      user
}

type user struct {
	ws    *websocket.Conn
	check bool
}

type Message struct {
	Data      string `json:"data"`
	YourColor string `json:"YourColor"`
	Message   string `json:"message"`
}

var upgrader = websocket.Upgrader{}

var OmokRoomData [max]OmokRoom

func main() {
	http.HandleFunc("/", index)
	http.Handle("/SOUND/", http.FileServer(http.FS(SOUND)))
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/ws", SocketHandler)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", html1)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", favicon)
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %v", err)
		return
	}
	RoomMatching(socket)
}

func RoomMatching(ws *websocket.Conn) {
	for {
		for i := 0; i < max; i++ {
			if OmokRoomData[i].uesr_1.check {
				if !OmokRoomData[i].uesr_2.check {
					OmokRoomData[i].uesr_2.check = true
					OmokRoomData[i].uesr_2.ws = ws
					OmokRoomData[i].MessageHandler()
					return
				}
			}
		}
		for i := 0; i < max; i++ {
			if !OmokRoomData[i].uesr_1.check {
				OmokRoomData[i].uesr_1.check = true
				OmokRoomData[i].uesr_1.ws = ws
				return
			}
		}
		time.Sleep(time.Second)
	}
}

func (room *OmokRoom) MessageHandler() {
	if !room.uesr_1.writing("", "black", "") || !room.uesr_2.writing("", "white", "") {
		room.reset()
		return
	}

	var i int
	var timeout bool
	var err bool

	for {
		i, timeout, err = reading(room.uesr_1.ws)
		if timeout {
			room.uesr_1.writing("", "", "패배(시간초과)")
			room.uesr_2.writing("", "", "승리(시간초과)")
			room.reset()
			return
		}
		if err {
			room.uesr_2.writing("", "", "승리(상대가 나감)")
			room.reset()
			return
		}
		if room.board_15x15[i] == emptied {
			room.board_15x15[i] = black
			if !room.uesr_2.writing(strconv.Itoa(i), "", "") || room.VictoryConfirm(i) {
				room.reset()
				return
			}
		} else {
			room.reset()
			return
		}

		i, timeout, err = reading(room.uesr_2.ws)
		if timeout {
			room.uesr_1.writing("", "", "승리(시간초과)")
			room.uesr_2.writing("", "", "패배(시간초과)")
			room.reset()
			return
		}
		if err {
			room.uesr_1.writing("", "", "승리(상대가 나감)")
			room.reset()
			return
		}
		if room.board_15x15[i] == emptied {
			room.board_15x15[i] = white
			if !room.uesr_1.writing(strconv.Itoa(i), "", "") || room.VictoryConfirm(i) {
				room.reset()
				return
			}
		} else {
			room.reset()
			return
		}

	}
}

func (room *OmokRoom) VictoryConfirm(index int) bool {
	directions := []int{15, 1, 16, 14}
	for _, direction := range directions {
		count := 1
		for i := 1; i <= 4; i++ {
			nextStoneIndex := (direction * i) + index
			if 0 <= nextStoneIndex && nextStoneIndex < 225 && room.board_15x15[nextStoneIndex] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		for i := -1; i >= -4; i-- {
			nextStoneIndex := (direction * i) + index
			if 0 <= nextStoneIndex && nextStoneIndex < 225 && room.board_15x15[nextStoneIndex] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		if count >= 5 {
			room.SendVictoryMessage(room.board_15x15[index])
			return true
		}
	}
	return false
}

func (room *OmokRoom) SendVictoryMessage(winnerColor uint8) {
	if winnerColor == black {
		room.uesr_1.writing("", "", "승리")
		room.uesr_2.writing("", "", "패배")

	} else {
		room.uesr_2.writing("", "", "승리")
		room.uesr_1.writing("", "", "패배")
	}
}

func reading(ws *websocket.Conn) (int, bool, bool) {
	timeoutDuration := 60 * time.Second
	ws.SetReadDeadline(time.Now().Add(timeoutDuration))

	_, m, err := ws.ReadMessage()
	if err != nil {
		log.Printf("conn.ReadMessage: %v", err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0, true, false
		}
		return 0, false, true
	}
	i, _ := strconv.Atoi(string(m))
	return i, false, false
}

func (user *user) writing(d, y, m string) bool {
	msg := Message{d, y, m}
	if err := user.ws.WriteJSON(msg); err != nil {
		log.Printf("conn.WriteMessage: %v", err)
		return false
	}
	return true
}

func (room *OmokRoom) reset() {
	room.uesr_1.check = false
	room.uesr_2.check = false
	room.board_15x15 = [225]uint8{}
	room.uesr_1.ws.Close()
	room.uesr_2.ws.Close()
}
