package main

import (
	"testing"
)

func TestGetCoordinates(t *testing.T) {
	level := new(LevelInfo)
	level.Tasks = append(level.Tasks, TaskInfo{false, "Рухайтесь за координатами <a href=\"google.navigation:q=49.46817, 23.43188\">49.46817, 23.43188</a>\r\n\r\nПо дорозі зніміть античіт: слово синьою фарбою на білому фоні отут <a href=\"google.navigation:q=49.473578, 23.470652\">49.473578, 23.470652</a>", "nil"})
	t.Logf("Getiing coordinates: %s", level.GetCoordinates())
	t.Errorf("Failed")
}
