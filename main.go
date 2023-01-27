package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

//go:embed HTML/1.html
var html1 string

const max = 100

const black uint8 = 1
const white uint8 = 2
const emptied uint8 = 0

type omok_room struct {
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
	Color     string `json:"color"`
	YourColor string `json:"YourColor"`
	Message   string `json:"message"`
}

var upgrader = websocket.Upgrader{}

var omok_data [max]omok_room

func main() {
	http.HandleFunc("/ws", socket_handler)
	http.HandleFunc("/omok", index)
	http.ListenAndServe(":8080", nil)
}

func socket_handler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %v", err)
		socket.Close()
		return
	}
	room_matching(socket)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", html1)
}

func room_matching(ws *websocket.Conn) {
	for {
		for i := 0; i < max; i++ {
			if !omok_data[i].uesr_1.check || !omok_data[i].uesr_2.check {
				if !omok_data[i].uesr_1.check {
					omok_data[i].uesr_1.check = true
					omok_data[i].uesr_1.ws = ws
					return
				}
				if !omok_data[i].uesr_2.check {
					omok_data[i].uesr_2.check = true
					omok_data[i].uesr_2.ws = ws
					omok_data[i].message_handler()
					return
				}
			}
		}
	}
}

func (room *omok_room) message_handler() {
	if room.uesr_1.writing("", "", "black", "") || room.uesr_2.writing("", "", "white", "") {
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
			if room.uesr_1.writing(string(m1), "black", "", "") || room.uesr_2.writing(string(m1), "black", "", "") || room.victory_Confirm(i1) {
				room.reset()
				return
			}
		}
		_, m2, err := room.uesr_2.ws.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			room.reset()
			return
		}
		i2, _ := strconv.Atoi(string(m2))
		if room.board_15x15[i2] == emptied {
			room.board_15x15[i2] = white
			if room.uesr_2.writing(string(m2), "white", "", "") || room.uesr_1.writing(string(m2), "white", "", "") || room.victory_Confirm(i2) {
				room.reset()
				return
			}
		}
	}
}

func (room *omok_room) victory_Confirm(index int) bool {
	var cases int
	for i := 1; i <= 4; i++ {
		count := 1
		switch i {
		case 1:
			cases = 15
		case 2:
			cases = 1

		case 3:
			cases = 16

		case 4:
			cases = 14
		}

		for i := 1; i <= 4; i++ {
			if 0 <= (cases*i)+index && (cases*i)+index < 225 && room.board_15x15[(cases*i)+index] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		for i := -1; i >= -4; i-- {
			if 0 <= (cases*i)+index && (cases*i)+index < 225 && room.board_15x15[(cases*i)+index] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		if count >= 5 {
			if room.board_15x15[index] == black {
				room.uesr_1.writing("", "", "", "승리")
				room.uesr_2.writing("", "", "", "패배")
			} else {
				room.uesr_1.writing("", "", "", "패배")
				room.uesr_2.writing("", "", "", "승리")
			}
			return true
		}
	}
	return false
}

func (user *user) writing(d, c, y, m string) bool {
	msg := Message{d, c, y, m}
	if err := user.ws.WriteJSON(msg); err != nil {
		log.Printf("conn.WriteMessage: %v", err)
		return true
	}
	return false
}

func (room *omok_room) reset() {
	room.uesr_1.check = false
	room.uesr_1.ws.Close()
	room.uesr_2.check = false
	room.uesr_2.ws.Close()
	room.board_15x15 = [225]uint8{}
}
