package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/bonya_bot/en"
)

// CoordinatesResponse represent the response that is sent to the user
type CoordinatesResponse struct {
	LevelNumber int8           `json:"level"`
	Coords      en.Coordinates `json:"coordinates"`
}

func getCoordinates(w http.ResponseWriter, r *http.Request, engine *en.API) {
	var (
		response = CoordinatesResponse{}
		buf      bytes.Buffer
	)

	log.Print("Get coordinates request accepted")

	if engine.CurrentLevel != nil {
		response.LevelNumber = engine.CurrentLevel.Number
		//log.Printf("%p", &en.CurrentLevel.Coords)
		_, response.Coords = en.ExtractCoordinates(engine.CurrentLevel.Tasks[0].TaskText)
		for _, hi := range engine.CurrentLevel.Helps {
			if hi.HelpText != "" {
				_, coords := en.ExtractCoordinates(hi.HelpText)
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, *en.API), en *en.API) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, en)
	}
}

func initHandlers(en *en.API) {
	log.Print("Adding enpoint handlers...")
	http.HandleFunc("/coords", makeHandler(getCoordinates, en))
}

func startServer(en *en.API) {
	initHandlers(en)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
