package main

import (
	"fmt"
	"net/http"
	"strings"
)

func Index(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[1:]
	if strings.HasSuffix(param, "/") {
		http.Redirect(w, r, "/"+strings.TrimSuffix(param, "/"), http.StatusPermanentRedirect)
		return
	}

	if param == "" {
		param = "index"
	}

	data, err := HTML.ReadFile(fmt.Sprintf("%s/%s.html", HTMLPath, param))
	if err == nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}

	data, err = HTML.ReadFile(fmt.Sprintf("%s/%s/index.html", HTMLPath, param))
	if err == nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}

	data, err = CONFIGS.ReadFile(fmt.Sprintf("%s/%s", ConfigPath, param))
	if err == nil {
		w.Write(data)
		return
	}

	serveErrorPage(w)
}

func serveErrorPage(w http.ResponseWriter) {
	data, err := CONFIGS.ReadFile("CONFIGS/err.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	w.Write(data)
}
