package translate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sbaglivi/TL-Pokedex/cache"
	"github.com/sbaglivi/TL-Pokedex/types"
)

type TranslationService struct {
	cache   cache.Cache
	baseURL *url.URL
	client  *http.Client
}

type Total struct {
	Total int
}

type Content struct {
	Translation string
	Text        string
	Translated  string
}

type TranslationResponse struct {
	Success  Total   `json:"success"`
	Contents Content `json:"contents"`
}

func NewTranslationService(cache cache.Cache, baseURL string, client *http.Client) (*TranslationService, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &TranslationService{
		cache,
		parsed,
		client,
	}, nil
}

func (ts *TranslationService) toURL(tsl types.Translation) string {
	rel, _ := url.Parse(string(tsl) + ".json")
	return ts.baseURL.ResolveReference(rel).String()
}

func (ts *TranslationService) translateWithAPI(s string, translation types.Translation) (*string, error) {
	url := ts.toURL(translation)
	body := map[string]string{"text": s}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("while serializing %s: %w", s, err)
	}
	resp, err := ts.client.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("while making translation request of type %s: %w", translation, err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var tslResponse TranslationResponse
	if err := json.Unmarshal(respBody, &tslResponse); err != nil {
		panic(err)
	}

	if tslResponse.Success.Total != 1 {
		return nil, errors.New("translation API response has success != 1")
	}
	return &tslResponse.Contents.Translated, nil
}

func (ts *TranslationService) Translate(key, value string, translation types.Translation) (*string, error) {
	if value == "" {
		return nil, fmt.Errorf("translation requested with key %s where value is empty", key)
	}

	key = key + "_translation"
	cached, exists := ts.cache.Get(key)
	if exists {
		return cached.(*string), nil
	}

	translated, err := ts.translateWithAPI(value, translation)
	if err != nil {
		return nil, err
	}

	ts.cache.Put(key, translated)
	return translated, nil
}
