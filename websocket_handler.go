package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func upgradeWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// upgrader.CheckOrigin = func(r *http.Request) bool {
	// 	return true
	// }

	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Connection upgrade failed: %v", err)
		http.Error(w, "Connection upgrade failed", http.StatusInternalServerError)
		return nil, err
	}

	sockets = append(sockets, socket)
	return socket, nil
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgradeWebSocketConnection(w, r)
	if err != nil {
		return
	}
	RoomMatching(socket)
}

func SpectatorHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgradeWebSocketConnection(w, r)
	if err != nil {
		return
	}
	BroadcastConnectionsCount()

	for {
		roomWithUser2Exists := false
		for _, room := range rooms {
			if room.user2.check {
				message := SpectatorMessage{room.board_15x15, nil, nil, room.user1.nickname, room.user2.nickname}
				if err := socket.WriteJSON(message); err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					handleSocketError(socket)
					return
				}

				room.spectators = append(room.spectators, socket)

				_, _, err := socket.ReadMessage()
				if err != nil {
					log.Printf("Error reading from WebSocket: %v", err)
					handleSocketError(socket)
					return
				}

				removeSocketFromSpectators(room, socket)
				roomWithUser2Exists = true
			}
		}

		if !roomWithUser2Exists {
			if IsWebSocketConnected(socket) {
				time.Sleep(5 * time.Second)
			} else {
				handleSocketError(socket)
				return
			}
		}
	}
}

func IsWebSocketConnected(conn *websocket.Conn) bool {
	log.Println("Checking WebSocket connection...")
	if err := conn.WriteJSON(map[string]interface{}{"type": WebSocketPingType}); err != nil {
		log.Printf("Failed to send Ping message: %v", err)
		return false
	}

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("Failed to set Read deadline: %v", err)
		return false
	}
	defer conn.SetReadDeadline(time.Time{})

	if _, pong, err := conn.ReadMessage(); err != nil || string(pong) != WebSocketPongType {
		log.Printf("Failed to receive Pong message: %v", err)
		return false
	}

	return true
}

func handleSocketError(socket *websocket.Conn) {
	removeWebSocketFromSockets(socket)
	socket.Close()
}

func removeSocketFromSpectators(room *OmokRoom, socket *websocket.Conn) {
	for i, r := range room.spectators {
		if r == socket {
			room.spectators = append(room.spectators[:i], room.spectators[i+1:]...)
			break
		}
	}
}

func removeWebSocketFromSockets(socket *websocket.Conn) {
	for i, r := range sockets {
		if r == socket {
			sockets = append(sockets[:i], sockets[i+1:]...)
			break
		}
	}
}
