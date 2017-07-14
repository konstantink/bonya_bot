package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type CoordinatesResponse struct {
	LevelNumber int8        `json:"level"`
	Coords      Coordinates `json:"coordinates"`
}

func getCoordinates(w http.ResponseWriter, r *http.Request, en *EnAPI) {
	var (
		response CoordinatesResponse = CoordinatesResponse{}
		buf      bytes.Buffer
	)

	log.Print("Get coordinates request accepted")

	if en.CurrentLevel != nil {
		response.LevelNumber = en.CurrentLevel.Number
		//log.Printf("%p", &en.CurrentLevel.Coords)
		_, response.Coords = ReplaceCoordinates(en.CurrentLevel.Tasks[0].TaskText)
		for _, hi := range en.CurrentLevel.Helps {
			if hi.HelpText != "" {
				_, coords := ReplaceCoordinates(hi.HelpText)
				response.Coords = append(response.Coords, coords...)
			}
		}
	}

	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		http.Error(w, "Something bad happened", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *EnAPI), en *EnAPI) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, en)
	}
}

func initHandlers(en *EnAPI) {
	log.Print("Adding enpoint handlers...")
	http.HandleFunc("/coords", makeHandler(getCoordinates, en))
}

func startServer(en *EnAPI) {
	initHandlers(en)
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", nil))
}
