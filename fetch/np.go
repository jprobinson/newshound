package fetch

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jasonmoo/toget"

	"github.com/jprobinson/newshound"
)

type npResp struct {
	NounPhrases map[string]string    `json:"noun_phrases"`
	Sentences   []newshound.Sentence `json:"sentences"`
	TopSentence string               `json:"top_sentence"`
}

func callNP(body []byte) (tags []string, sentences []newshound.Sentence, topSentence string, err error) {
	var resp *http.Response
	resp, err = toget.Post("http://localhost:1029/", bytes.NewReader(body), time.Second*5)
	if err != nil {
		log.Print("unable to hit np_extractor: ", err)
		return
	}

	var npR npResp
	if err = json.NewDecoder(resp.Body).Decode(&npR); err != nil {
		return
	}

	for tag, _ := range npR.NounPhrases {
		tags = append(tags, tag)
	}
	sentences = npR.Sentences
	topSentence = npR.TopSentence
	return
}
