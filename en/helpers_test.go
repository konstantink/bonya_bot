package main

import (
	"testing"
)

type blockTypeTestPair struct {
	code  int8
	block string
}

func TestBlockTypeToString(t *testing.T) {
	var testpairs = []blockTypeTestPair{
		{0, "Игрок"},
		{1, "Игрок"},
		{2, "Команда"},
		{-1, "Команда"},
		{10, "Команда"},
	}

	for _, pair := range testpairs {
		if res := BlockTypeToString(pair.code); res != pair.block {
			t.Errorf("For %d expected %q, got %q", pair.code, pair.block, res)
		}
	}
}

// =================================================================================== //

func TestReplaceCoordinates(t *testing.T) {
	var examples = []struct {
		input    string
		expected string
		coords   Coordinates
	}{
		{`<a href="geo:49.976136, 36.267256">49.976136, 36.267256</a>`, "49.976136, 36.267256",
			Coordinates{{Lon: 49.976136, Lat: 36.267256, OriginalString: "49.976136, 36.267256"}}},
		{`<a href="https://www.google.com.ua/maps/@50.0363257,36.2120039,19z" target="blank">50.036435 36.211914</a>`, "50.036435 36.211914",
			Coordinates{{Lon: 50.036326, Lat: 36.212004, OriginalString: "50.036435 36.211914"}}},
	}

	//"49.976136, 36.267256"
	for _, ex := range examples {
		if res, _ := ReplaceCoordinates(ex.input); res != ex.expected {
			t.Errorf("For %q\nExpected %q\nGot      %q",
				ex.input, ex.expected, res)
		}
	}
}
