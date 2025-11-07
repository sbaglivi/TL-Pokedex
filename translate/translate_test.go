package translate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sbaglivi/TL-Pokedex/cache"
	"github.com/sbaglivi/TL-Pokedex/types"
)

func TestTranslate(t *testing.T) {
	to_translate := "It's a good morning"
	correct := "A good morning it is"
	resp := TranslationResponse{
		Contents: Content{
			Translation: "yoda",
			Text:        to_translate,
			Translated:  correct,
		},
		Success: Total{
			Total: 1,
		},
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic("failed to marshal translation response")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(bytes)
	}))
	defer srv.Close()

	cache := cache.NewLRU(10)
	svc, err := NewTranslationService(cache, srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("failed to instantiate translation service: %v", err)
	}
	ctx := context.Background()
	translated, err := svc.Translate(ctx, "pikachu", to_translate, types.Yoda)
	if err != nil {
		t.Fatalf("translation failed with error: %v", err)
	}
	if *translated != correct {
		t.Fatalf("translation failed: expected %s received %s", correct, *translated)
	}

}

func TestTranslationURL(t *testing.T) {
	cache := cache.NewLRU(10)
	baseURL := "http://fakeapi.com"
	svc, err := NewTranslationService(cache, baseURL, http.DefaultClient)
	if err != nil {
		t.Errorf("failed to instantiate translation service: %v", err)
	}
	result := svc.toURL(types.Yoda)
	expected := "http://fakeapi.com/yoda.json"
	if result != expected {
		t.Errorf("URL built %s does not match expected %s", result, expected)
	}
	expected = "http://fakeapi.com/shakespeare.json"
	result = svc.toURL(types.Shakespeare)
	if result != expected {
		t.Errorf("URL built %s does not match expected %s", result, expected)
	}

	baseURL = "http://fakeapi.com/"
	svc, err = NewTranslationService(cache, baseURL, http.DefaultClient)
	if err != nil {
		t.Errorf("failed to instantiate translation service: %v", err)
	}
	result = svc.toURL(types.Shakespeare)
	if result != expected {
		t.Errorf("URL built %s does not match expected %s", result, expected)
	}
}
