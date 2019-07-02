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

func decodeAlert(ctx context.Context, r *http.Request) (interface{}, error) {
	var msg psmessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to decode request. dismissing message.")
		return nil, kit.NewJSONStatusResponse("bad request", http.StatusOK)
	}

	var alert newshound.NewsAlertLite
	err = gob.NewDecoder(bytes.NewBuffer(msg.Message.Data)).Decode(&alert)
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to ungob payload. dismissing message.")
		return nil, kit.NewJSONStatusResponse("bad request", http.StatusOK)
	}

	return alert, nil
}

func (s *service) postAlert(ctx context.Context, r interface{}) (interface{}, error) {
	alert := r.(newshound.NewsAlertLite)

	for _, barker := range s.alertsOut {
		err := barker.Bark(alert)
		if err != nil {
			kit.LogErrorMsg(ctx, err, "problems barking about alert")
		}
	}

	return "OK", nil
}
