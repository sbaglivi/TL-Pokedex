package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sbaglivi/TL-Pokedex/types"
	"github.com/sbaglivi/TL-Pokedex/utils"
	"golang.org/x/sync/singleflight"
)

type TranslationService struct {
	cache                types.Cache
	baseURL              *url.URL
	client               *http.Client
	group                singleflight.Group
	translateWithAPIfunc func(context.Context, string, types.Translation) (*string, error)
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

type TranslationError struct {
	Message string `json:"message"`
}

type TranslationErrorResponse struct {
	Error TranslationError `json:"error"`
}

func NewTranslationService(cache types.Cache, baseURL string, client *http.Client) (*TranslationService, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	svc := TranslationService{
		cache:   cache,
		baseURL: parsed,
		client:  client,
	}
	svc.translateWithAPIfunc = svc.translateWithAPI
	return &svc, nil
}

func (ts *TranslationService) toURL(tsl types.Translation) string {
	rel, _ := url.Parse(string(tsl) + ".json")
	return ts.baseURL.ResolveReference(rel).String()
}

func getErrorMessage(body []byte) string {
	var errorResponse TranslationErrorResponse
	err := json.Unmarshal(body, &errorResponse)
	if err != nil {
		return string(body[:1024])
	}

	return errorResponse.Error.Message
}

func (ts *TranslationService) groupedTranslateWithAPI(ctx context.Context, s string, translation types.Translation) (*string, error) {
	// don't expect same desc for pokemons with different translation methods but just being safe
	key := fmt.Sprintf("%s-%s", s, string(translation))
	translated, err, shared := ts.group.Do(key, func() (interface{}, error) {
		return ts.translateWithAPIfunc(ctx, s, translation)
	})

	if shared {
		slog.Debug("shared translation API request", "key", key)
	}

	return translated.(*string), err
}

func (ts *TranslationService) translateWithAPI(ctx context.Context, s string, translation types.Translation) (*string, error) {
	url := ts.toURL(translation)
	body := map[string]string{"text": s}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w while serializing [%s] for translation request: %v", types.ErrGeneric, s, err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("%w while preparing POST for url %s to translate [%s]: %v", types.ErrGeneric, url, s, err)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w while making translation request of type %s for [%s]: %v", types.ErrGeneric, translation, s, err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w while reading response body for request of type %s: %v", types.ErrGeneric, translation, err)
	}

	if resp.StatusCode != http.StatusOK {
		detail := getErrorMessage(respBody)
		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Info("translation API rate limit hit", "detail", detail)
			return nil, types.ErrTooManyRequests
		}

		return nil, fmt.Errorf("%w unexpected status %d from upstream while requesting translation of type %s for [%s]: %s", types.ErrGeneric, resp.StatusCode, string(translation), s, detail)
	}

	var tslResponse TranslationResponse
	if err := json.Unmarshal(respBody, &tslResponse); err != nil {
		return nil, fmt.Errorf("%w while unmarshaling response for translation of type %s: %v", types.ErrGeneric, translation, err)
	}

	if tslResponse.Success.Total != 1 {
		return nil, fmt.Errorf("%w response for translation of type %s for [%s] has success.total != 1", types.ErrGeneric, translation, s)
	}
	cleaned := utils.RemoveWhitespace(tslResponse.Contents.Translated)
	return &cleaned, nil
}

func (ts *TranslationService) Translate(ctx context.Context, key, value string, translation types.Translation) (*string, error) {
	if value == "" {
		slog.Debug(fmt.Sprintf("translation requested with key %s where value is empty", key))
		return &value, nil
	}

	key = key + "_translation"
	cached, exists := ts.cache.Get(key)
	if exists {
		return cached.(*string), nil
	}

	translated, err := ts.groupedTranslateWithAPI(ctx, value, translation)
	if err != nil {
		return nil, err
	}

	ts.cache.Put(key, translated)
	return translated, nil
}
