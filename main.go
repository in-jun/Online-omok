package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", Index)
	http.Handle("/IMAGE/", http.FileServer(http.FS(IMAGE)))
	http.Handle("/SOUND/", http.FileServer(http.FS(SOUND)))
	http.HandleFunc("/game", SocketHandler)
	http.HandleFunc("/spectator", SpectatorHandler)
	http.ListenAndServe(serverAddress, nil)
}
