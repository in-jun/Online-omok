package main

import (
	"log"
	"net"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

func RoomMatching(ws *websocket.Conn) {
	log.Println("Waiting for room matching...")
	connectionsCount++
	BroadcastConnectionsCount()

	_, nickname, err := ws.ReadMessage()
	if err != nil || utf8.RuneCountInString(string(nickname)) > MaxNicknameLength {
		log.Printf("Error reading nickname from WebSocket: %v", err)
		handleFailedRoomMatching()
		return
	}

	for _, room := range rooms {
		if room.user1.check && !room.user2.check {
			if IsWebSocketConnected(room.user1.ws) {
				if IsWebSocketConnected(ws) {
					room.user2.check = true
					room.user2.ws = ws
					room.user2.nickname = string(nickname)
					log.Println("User 2 joined room")
					room.MessageHandler()
				} else {
					handleFailedRoomMatching()
				}
				return
			} else {
				room.reset()
			}
		}

	}
	newRoom := &OmokRoom{}
	newRoom.user1.check = true
	newRoom.user1.ws = ws
	newRoom.user1.nickname = string(nickname)
	rooms = append(rooms, newRoom)
	log.Println("User 1 created a new room")
}

func (room *OmokRoom) MessageHandler() {
	log.Println("Starting the game in the room...")
	err1 := room.user1.ws.WriteJSON(Message{nil, "black", nil, nil, room.user2.nickname})
	err2 := room.user2.ws.WriteJSON(Message{nil, "white", nil, nil, room.user1.nickname})
	if err1 != nil || err2 != nil {
		log.Println("Failed to set up the game. Resetting the room.")
		room.reset()
		return
	}

	var i int
	var timeout bool
	var err bool

	for {
		i, timeout, err = reading(room.user1.ws)
		if timeout {
			handleGameTimeout(room, StatusUser1Timeout, StatusUser2Timeout)
			log.Println("User 1 timeout. User 2 wins. Resetting the room.")
			return
		}
		if err {
			room.user2.ws.WriteJSON(Message{nil, nil, 4, nil, nil})
			room.reset()
			log.Println("Error reading from User 1. Resetting the room.")
			return
		}
		if room.isValidMove(i) {
			room.board_15x15[i] = black
			err := room.user2.ws.WriteJSON(Message{i, nil, nil, nil, nil})
			if err != nil || room.VictoryConfirm(i) {
				room.reset()
				return
			}
			room.broadcastToSpectators(i, black)
		} else {
			room.reset()
			return
		}

		i, timeout, err = reading(room.user2.ws)
		if timeout {
			handleGameTimeout(room, StatusUser2Timeout, StatusUser1Timeout)
			log.Println("User 2 timeout. User 1 wins. Resetting the room.")
			return
		}
		if err {
			room.user1.ws.WriteJSON(Message{nil, nil, StatusErrorReading, nil, nil})
			room.reset()
			log.Println("Error reading from User 2. Resetting the room.")
			return
		}
		if room.isValidMove(i) {
			room.board_15x15[i] = white
			err := room.user1.ws.WriteJSON(Message{i, nil, nil, nil, nil})
			if err != nil || room.VictoryConfirm(i) {
				room.reset()
				return
			}
			room.broadcastToSpectators(i, white)
		} else {
			room.reset()
			return
		}
	}
}

func (room *OmokRoom) isValidMove(index int) bool {
	return 0 <= index && index < 225 && room.board_15x15[index] == emptied
}

func (room *OmokRoom) broadcastToSpectators(n int, color uint8) {
	for _, ws := range room.spectators {
		ws.WriteJSON(SpectatorMessage{nil, n, color, nil, nil})
	}
}

func (room *OmokRoom) VictoryConfirm(index int) bool {
	directions := []int{BoardSize, 1, BoardSize + 1, BoardSize - 1}
	for _, direction := range directions {
		count := 1
		for i := 1; i <= WinningCount; i++ {
			nextStoneIndex := (direction * i) + index
			if 0 <= nextStoneIndex && nextStoneIndex < TotalCells && room.board_15x15[nextStoneIndex] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		for i := -1; i >= -WinningCount; i-- {
			nextStoneIndex := (direction * i) + index
			if 0 <= nextStoneIndex && nextStoneIndex < TotalCells && room.board_15x15[nextStoneIndex] == room.board_15x15[index] {
				count++
			} else {
				break
			}
		}
		if count == WinningCount {
			room.SendVictoryMessage(room.board_15x15[index])
			return true
		}
	}
	return false
}

func (room *OmokRoom) SendVictoryMessage(winnerColor uint8) {
	if winnerColor == black {
		room.user1.ws.WriteJSON(Message{nil, nil, 0, nil, nil})
		room.user2.ws.WriteJSON(Message{nil, nil, 1, nil, nil})

	} else {
		room.user2.ws.WriteJSON(Message{nil, nil, 0, nil, nil})
		room.user1.ws.WriteJSON(Message{nil, nil, 1, nil, nil})
	}
}

func reading(ws *websocket.Conn) (int, bool, bool) {
	log.Println("Reading from WebSocket...")
	ws.SetReadDeadline(time.Now().Add(TimeoutDuration))

	_, m, err := ws.ReadMessage()
	if err != nil {
		log.Printf("Error reading from WebSocket: %v", err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0, true, false
		}
		return 0, false, true
	}
	i, _ := strconv.Atoi(string(m))
	return i, false, false
}

func (room *OmokRoom) reset() {
	log.Println("Resetting the room...")

	removeWebSocketFromSockets(room.user1.ws)
	removeWebSocketFromSockets(room.user2.ws)

	if room.user1.ws != nil {
		room.user1.ws.Close()
		connectionsCount--
	}
	if room.user2.ws != nil {
		room.user2.ws.Close()
		connectionsCount--
	}

	removeRoomFromRooms(room)

	BroadcastConnectionsCount()
}

func handleFailedRoomMatching() {
	connectionsCount--
	BroadcastConnectionsCount()
}

func handleGameTimeout(room *OmokRoom, winnerStatus, loserStatus int) {
	room.user1.ws.WriteJSON(Message{nil, nil, winnerStatus, nil, nil})
	room.user2.ws.WriteJSON(Message{nil, nil, loserStatus, nil, nil})
	room.reset()
	log.Printf("User timeout. Winner wins. Resetting the room.")
}

func removeRoomFromRooms(room *OmokRoom) {
	for i, r := range rooms {
		if r == room {
			rooms = append(rooms[:i], rooms[i+1:]...)
			break
		}
	}
}

func BroadcastConnectionsCount() {
	for _, socket := range sockets {
		socket.WriteJSON(Message{nil, nil, nil, connectionsCount, nil})
	}
}
