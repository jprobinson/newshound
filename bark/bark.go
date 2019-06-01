package bark

import "github.com/jprobinson/newshound"

type AlertBarker interface {
	Bark(alert newshound.NewsAlertLite) error
}

type AlertBarkerFunc func(alert newshound.NewsAlertLite) error

func (a AlertBarkerFunc) Bark(alert newshound.NewsAlertLite) error {
	return a(alert)
}

type EventBarker interface {
	Bark(alert newshound.NewsEvent) error
}

type EventBarkerFunc func(event newshound.NewsEvent) error

func (e EventBarkerFunc) Bark(event newshound.NewsEvent) error {
	return e(event)
}

var SenderColors = map[string]string{
	"cnn":                      "#B60002",
	"foxnews.com":              "#234E6C",
	"foxbusiness.com":          "#343434",
	"nbcnews.com":              "#343434",
	"nbc":                      "#343434",
	"nytimes.com":              "#1A1A1A",
	"the new york times":       "#1A1A1A",
	"the washington post":      "#222",
	"wsj.com":                  "#444242",
	"the wall street journal.": "#444242",
	"the wall street journal":  "#444242",
	"politico":                 "#256396",
	"los angeles times":        "#000",
	"cbs":                      "#313943",
	"abc":                      "#1b6295",
	"usatoday.com":             "#1877B6",
	"usatoday":                 "#1877B6",
	"usa today":                "#1877B6",
	"yahoo":                    "#7B0099",
	"ft":                       "#FFF1E0",
	"bbc":                      "#c00000",
	"npr":                      "#5f82be",
	"time":                     "#e90606",
	"bloomberg.com":            "#110c09",
}
