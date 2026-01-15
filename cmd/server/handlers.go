package main

import (
	"encoding/json"
	"fmt"

	"github.com/frogonabike/peril/internal/routing"
)

func handleGameLog(log []byte) AckType {
	defer fmt.Print("> ")
	var gamelog routing.GameLog
	err := json.Unmarshal(log, &gamelog)
	return Ack
}
