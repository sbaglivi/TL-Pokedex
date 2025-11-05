package pokemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sbaglivi/TL-Pokedex/cache"
	"github.com/sbaglivi/TL-Pokedex/translate"
)

func TestAPIPokemonToInternal(t *testing.T) {
	apiPokemon := APIPokemon{
		IsLegendary: true,
		Name:        "Groudon",
		APIHabitat: NameAndURL{
			Name: "Lava",
		},
		FlavorTextEntries: []FlavorTextEntry{
			{
				FlavorText: "   Test\r\nof a weird     description",
			},
		},
	}

	if !reflect.DeepEqual(apiPokemon.toInternal(),
		Pokemon{
			Name:        "Groudon",
			IsLegendary: true,
			Desc:        "Test of a weird description",
			Habitat:     "Lava",
		}) {
		t.Fatalf("unexpected converted pokemon: %#v", apiPokemon.toInternal())
	}
}

func TestGetTranslatedPokemon(t *testing.T) {
	cache := cache.NewLRU(10)
	translationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"contents":{"translation":"yoda","text":"It's a good morning","translated":"A good morning it is"},"success":{"total": 1}}`))
	}))
	defer translationServer.Close()
	translationService, err := translate.NewTranslationService(cache, translationServer.URL, translationServer.Client())
	if err != nil {
		t.Fatalf("while creating translate service: %v", err)
	}
	apiPokemon := APIPokemon{
		IsLegendary: true,
		Name:        "Groudon",
		APIHabitat: NameAndURL{
			Name: "Lava",
		},
		FlavorTextEntries: []FlavorTextEntry{
			{
				FlavorText: "   Test\r\nof a weird     description",
			},
		},
	}

	bytes, _ := json.Marshal(apiPokemon)
	pkmnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(bytes)
	}))
	defer pkmnServer.Close()

	pkmnService, err := NewPokemonService(cache, translationService, pkmnServer.URL, pkmnServer.Client())
	if err != nil {
		t.Fatalf("failed to create pokemonService: %v", err)
	}
	pkmn, err := pkmnService.GetPokemon("groudon", false)
	if err != nil {
		t.Fatalf("failed to GetPokemon('groudon', false): %v", err)
	}
	expect := "Test of a weird description"
	if pkmn.Desc != expect {
		t.Fatalf("GetPokemon('groudon') returned %s expected %s", pkmn.Desc, expect)
	}

	pkmn, err = pkmnService.GetPokemon("groudon", true)
	if err != nil {
		t.Fatalf("failed to GetPokemon('groudon', true): %v", err)
	}
	expect = "A good morning it is"
	if pkmn.Desc != expect {
		t.Errorf("GetPokemon('groudon') returned %s expected %s", pkmn.Desc, expect)
	}

}

func TestPokemonCachingBehavior(t *testing.T) {
	cache := cache.NewLRU(10)

	var pkmnCalls, translationCalls int

	translationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		translationCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"contents":{"translation":"yoda","text":"It's a good morning","translated":"A good morning it is"},"success":{"total": 1}}`))
	}))
	defer translationServer.Close()

	translationService, err := translate.NewTranslationService(cache, translationServer.URL, translationServer.Client())
	if err != nil {
		t.Fatalf("creating translate service: %v", err)
	}

	apiPokemon := APIPokemon{
		IsLegendary: true,
		Name:        "Groudon",
		APIHabitat: NameAndURL{
			Name: "Lava",
		},
		FlavorTextEntries: []FlavorTextEntry{
			{FlavorText: "   Test\r\nof a weird     description"},
		},
	}
	bytes, _ := json.Marshal(apiPokemon)

	pkmnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkmnCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(bytes)
	}))
	defer pkmnServer.Close()

	pkmnService, err := NewPokemonService(cache, translationService, pkmnServer.URL, pkmnServer.Client())
	if err != nil {
		t.Fatalf("creating pokemon service: %v", err)
	}

	pkmn, err := pkmnService.GetPokemon("groudon", true)
	if err != nil {
		t.Fatalf("GetPokemon('groudon', true) failed: %v", err)
	}
	if pkmn.Desc != "A good morning it is" {
		t.Errorf("unexpected description: %q", pkmn.Desc)
	}
	if pkmnCalls != 1 {
		t.Errorf("expected 1 pokemon API call, got %d", pkmnCalls)
	}
	if translationCalls != 1 {
		t.Errorf("expected 1 translation API call, got %d", translationCalls)
	}

	// ---- second call (same params): should hit cache only ----
	pkmn, err = pkmnService.GetPokemon("groudon", true)
	if err != nil {
		t.Fatalf("GetPokemon('groudon', true) failed: %v", err)
	}
	if pkmnCalls != 1 {
		t.Errorf("pokemon API should not be called again, got %d", pkmnCalls)
	}
	if translationCalls != 1 {
		t.Errorf("translation API should not be called again, got %d", translationCalls)
	}

	// ---- third call (no translation): should use cached Pokémon, skip translation ----
	pkmn, err = pkmnService.GetPokemon("groudon", false)
	if err != nil {
		t.Fatalf("GetPokemon('groudon', false) failed: %v", err)
	}
	if pkmnCalls != 1 {
		t.Errorf("pokemon API should still not be called again, got %d", pkmnCalls)
	}
	if translationCalls != 1 {
		t.Errorf("translation API should still not be called again, got %d", translationCalls)
	}
}
