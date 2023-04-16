package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed HTML/1.html
var html1 string

//go:embed IMAGE/favicon.ico
var favicon string

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
	for {
		_, m1, err := room.uesr_1.ws.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			room.reset()
			return
		}
		i1, _ := strconv.Atoi(string(m1))
		if room.board_15x15[i1] == emptied {
			room.board_15x15[i1] = black
			if !room.uesr_2.writing(string(m1), "", "") || room.VictoryConfirm(i1) {
				room.reset()
				return
			}
		}
		_, m2, err := room.uesr_2.ws.ReadMessage()
		if err != nil {
			room.reset()
			log.Printf("conn.ReadMessage: %v", err)
			return
		}
		i2, _ := strconv.Atoi(string(m2))
		if room.board_15x15[i2] == emptied {
			room.board_15x15[i2] = white
			if !room.uesr_1.writing(string(m2), "", "") || room.VictoryConfirm(i2) {
				room.reset()
				return
			}
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
