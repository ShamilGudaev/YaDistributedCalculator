package events

import (
	"encoding/json"
	"fmt"

	"github.com/olebedev/emitter"
)

var EventsEmitter = &emitter.Emitter{}

func SendEventToClients(name string, data interface{}) error {
	data0, err := json.Marshal(data)
	if err != nil {
		return err
	}
	EventsEmitter.Emit("client", fmt.Sprintf("event: %s\ndata: %s\n\n", name, string(data0)))
	return nil
}
