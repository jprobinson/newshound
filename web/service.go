package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"google.golang.org/appengine/log"

	"github.com/nytimes/gizmo/server"
	"github.com/tardisgo/tardisgo/goroot/haxe/go1.4/src/html/template"
)

type service struct {
}

func (s *service) Middleware(h http.Handler) http.Handler {
	return h
}

func (s *service) JSONMiddleware(e server.JSONEndpoint) server.JSONEndpoint {
	return e
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/": s.homePageHandler,
		//		"/calendar": s.calendarHandler,
		//		"/reports":  s.reportsHandler,
	}
}

func (s *service) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	panic("not implemented")
}

func (s *service) Prefix() string {
	return ""
}

func (s *service) homePageHandler(w http.ResponseWriter, r *http.Request) {

}

func encodeHomePage(ctx context.Context, w http.ResponseWriter) error {
	r := res.(*playResponse)
	tmpl, err := template.New("game").ParseFiles("./static/play.html")
	if err != nil {
		log.Criticalf(ctx, "unable to parse index.html: %s", err)
		return err
	}
	usrData, _ := json.Marshal(User{r.UID})
	puzData, _ := json.Marshal(r.Puzzle)
	stateData, _ := json.Marshal(r.State)
	data := struct {
		User         template.JS
		Token        string
		GameURI      string
		Puzzle       template.JS
		SVG          template.HTML
		State        template.JS
		Code         string
		FBASE        string
		FBASE_SENDER string
		FBASE_API    string
	}{template.JS(string(usrData)), r.Token, r.URI, template.JS(string(puzData)),
		template.HTML(r.Puzzle.Body[0].Board), template.JS(string(stateData)), r.Code,
		os.Getenv("FBASE"), os.Getenv("FBASE_SENDER"), os.Getenv("FBASE_API")}
	return tmpl.ExecuteTemplate(w, "play.html", &data)
}
