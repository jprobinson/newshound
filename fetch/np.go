package fetch

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

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
	resp.Body.Close()

	// normalize the results
	nrmlzr := normalizer()
	for tag, _ := range npR.NounPhrases {
		tags = append(tags, normalize(nrmlzr, tag))
	}
	for _, s := range npR.Sentences {
		for i, p := range s.Phrases {
			s.Phrases[i] = normalize(nrmlzr, p)
		}
		sentences = append(sentences, s)
	}
	topSentence = npR.TopSentence
	return
}

func normalizer() transform.Transformer {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	return transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
}

// http://blog.golang.org/normalization#TOC_10.
func normalize(normalizer transform.Transformer, in string) string {
	result, _, err := transform.String(normalizer, in)
	if err != nil {
		log.Printf("unable to transform text:\n\n%q\nerr: %s", in, err)
		return in
	}
	// replace non-breaking spaces!
	return strings.Replace(result, "\u00a0", " ", -1)
}
