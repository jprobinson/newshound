package fetch

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jprobinson/newshound"
)

type npResp struct {
	NounPhrases map[string]string    `json:"noun_phrases"`
	Sentences   []newshound.Sentence `json:"sentences"`
	TopSentence string               `json:"top_sentence"`
}

func callNP(body []byte) (tags []string, sentences []newshound.Sentence, topSentence string, err error) {
	var resp *http.Response
	resp, err = http.Post("http://127.0.0.1:1029/", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Print("unable to hit np_extractor: ", err)
		return
	}
	defer resp.Body.Close()

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
