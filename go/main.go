package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/delaneyj/datastar"
)

var (
	backendData = Store{}
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("GET /", mainHandler)
	router.HandleFunc("PUT /put", putHandler)
	router.HandleFunc("GET /get", getHandler)
	router.HandleFunc("GET /feed", feedHandler)

	server := http.Server{Addr: ":4000", Handler: router}
	server.ListenAndServe()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	Page().Render(ctx, w)
}

type Store struct {
	Input string `json:"input"`
	Show  bool   `json:"show"`
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	err := datastar.BodyUnmarshal(r, &backendData)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	frag := datastar.GET("<div id=\"output\">Your input: %s, is %d long.</div>", backendData.Input, len(backendData.Input))

	sse := datastar.NewSSE(w, r)
	datastar.ConsoleInfoF(sse, "Backend State: %+v.", backendData)
	datastar.RenderFragmentString(sse, frag)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	frag := datastar.GET("<div id=\"output2\">%+v</div>", backendData)
	datastar.RenderFragmentString(sse, frag)

	frag = `<div id="output3">Check this out!</div>`

	datastar.RenderFragmentString(sse, frag, datastar.WithMergePrepend(), datastar.WithQuerySelectorID("main"))
}

func feedHandler(w http.ResponseWriter, r *http.Request) {

	ticker := time.NewTicker(time.Second)

	sse := datastar.NewSSE(w, r)

	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			token, err := generateToken()

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			frag := FeedFragment(token)
			datastar.RenderFragmentTempl(sse, frag)
		}
	}
}

func generateToken() (string, error) {
	token := make([]byte, 8)
	_, err := rand.Read(token)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}
