package bark

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/jprobinson/newshound"
)

func decodeEvent(ctx context.Context, r *http.Request) (interface{}, error) {
	var msg psmessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to decode request. dismissing message.")
		return nil, kit.NewJSONStatusResponse("bad request", http.StatusOK)
	}

	var event newshound.NewsEvent
	err = gob.NewDecoder(bytes.NewBuffer(msg.Message.Data)).Decode(&event)
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to ungob payload. dismissing message.")
		return nil, kit.NewJSONStatusResponse("bad request", http.StatusOK)
	}

	return event, nil
}

func (s *service) postEvent(ctx context.Context, r interface{}) (interface{}, error) {
	event := r.(newshound.NewsEvent)

	for _, barker := range s.eventsOut {
		err := barker.Bark(event)
		if err != nil {
			kit.LogErrorMsg(ctx, err, "problems barking about event")
		}
	}

	return "OK", nil
}
