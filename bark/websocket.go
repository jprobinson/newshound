package bark

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jprobinson/newshound"
	"github.com/m4rw3r/uuid"
)

func AddWebSocketBarker(d *Distributor, port int, alerts, events bool) {
	w := WebSocketBarker{
		port:    port,
		sockets: map[string]chan interface{}{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/ws", w.ServeWS)
	go func() {
		log.Printf("listening on %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			log.Fatal(err)
		}
	}()

	if alerts {
		d.AddAlertBarker(AlertBarkerFunc(w.BarkAlert))
	}
	if events {
		d.AddEventBarker(EventBarkerFunc(w.BarkEvent))
	}
}

type WebSocketBarker struct {
	port int

	muSockets sync.RWMutex
	sockets   map[string]chan interface{}
}

var ErrTooManyConns = errors.New("too many concurrent connections. please try again later.")

func (w *WebSocketBarker) NewSocket() (alerts chan interface{}, quit chan struct{}, err error) {
	id, _ := uuid.V4()
	key := id.String()
	alerts = make(chan interface{}, 500)
	quit = make(chan struct{}, 1)

	w.muSockets.Lock()
	if len(w.sockets) > 500 {
		return alerts, quit, ErrTooManyConns
	}
	w.sockets[key] = alerts
	log.Printf("added socket %s", key)
	w.muSockets.Unlock()

	go func() {
		<-quit
		w.muSockets.Lock()
		delete(w.sockets, key)
		log.Printf("deleted socket %s", key)
		w.muSockets.Unlock()
	}()
	return alerts, quit, nil
}

func (w *WebSocketBarker) BarkAlert(alert newshound.NewsAlertLite) error {
	w.muSockets.RLock()
	for key, socket := range w.sockets {
		socket <- alert
		log.Printf("alerted socket %s:%s - %s", key, alert.Sender, alert.Subject)
	}
	w.muSockets.RUnlock()
	return nil
}

func (w *WebSocketBarker) BarkEvent(event newshound.NewsEvent) error {
	w.muSockets.RLock()
	for key, socket := range w.sockets {
		socket <- event
		log.Printf("alerted socket of event %s:%s", key, event.ID)
	}
	w.muSockets.RUnlock()
	return nil
}

func (wb *WebSocketBarker) ServeWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Print("websocket err: ", err)
		}
		return
	}
	var (
		alerts     chan interface{}
		quit       chan struct{}
		noopTicker = time.NewTicker(time.Second * 5)
	)
	alerts, quit, err = wb.NewSocket()
	if err != nil {
		http.Error(w, err.Error(), 503)
		return
	}
	defer func() {
		quit <- struct{}{}
		ws.Close()
		noopTicker.Stop()
	}()

	for {
		select {
		case alert := <-alerts:
			ws.SetWriteDeadline(time.Now().Add(time.Second * 30))
			if err := ws.WriteJSON(alert); err != nil {
				log.Print("websocket json err: ", err)
				return
			}
		case <-noopTicker.C:
			ws.SetWriteDeadline(time.Now().Add(time.Second * 30))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Print("websocket noop err: ", err)
				return
			}
		}
	}
}

var (
	homeTempl = template.Must(template.New("").Parse(homeHTML))
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var v = struct {
		Host string
	}{
		r.Host,
	}
	homeTempl.Execute(w, &v)
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Newshound Bark WS</title>
    </head>
    <body>
		<h1>Welcome to the Newshound barkd Web Socket!</h1>
		<div id="events">
		</div>
		<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
        <script type="text/javascript">
            (function() {
                var conn = new WebSocket("ws://{{.Host}}/ws");
                conn.onmessage = function(evt) {
                    var evts = $("#events");
                    evts.append("<p>"+evt.data+"</p>");
                }
            })();
        </script>
    </body>
</html>
`
