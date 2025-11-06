package pokemon

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/sbaglivi/TL-Pokedex/types"
)

type NameAndURL struct {
	Name string `json:"name"`
}

type FlavorTextEntry struct {
	FlavorText string `json:"flavor_text"`
}

type APIPokemon struct {
	IsLegendary       bool              `json:"is_legendary"`
	Name              string            `json:"name"`
	APIHabitat        NameAndURL        `json:"habitat"`
	FlavorTextEntries []FlavorTextEntry `json:"flavor_text_entries"`
}

type Pokemon struct {
	IsLegendary bool   `json:"is_legendary"`
	Name        string `json:"name"`
	Habitat     string `json:"habitat"`
	Desc        string `json:"desc"`
}

type Translator interface {
	Translate(string, string, types.Translation) (*string, error)
}

type PokemonService struct {
	cache      types.Cache
	translator Translator
	baseURL    *url.URL
	client     *http.Client
}

func NewPokemonService(cache types.Cache, translator Translator, baseURL string, client *http.Client) (*PokemonService, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &PokemonService{
		cache:      cache,
		translator: translator,
		baseURL:    parsed,
		client:     client,
	}, nil
}

var re = regexp.MustCompile(`\s+`)

func removeWhitespace(s string) string {
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func normalize(s string) string {
	return removeWhitespace(strings.ToLower(s))
}

func (pkmn *APIPokemon) toInternal() Pokemon {
	desc := ""
	if len(pkmn.FlavorTextEntries) > 0 {
		desc = removeWhitespace(pkmn.FlavorTextEntries[0].FlavorText)
	}
	return Pokemon{
		IsLegendary: pkmn.IsLegendary,
		Name:        pkmn.Name,
		Habitat:     pkmn.APIHabitat.Name,
		Desc:        desc,
	}
}

func (ps *PokemonService) getPokemonURL(name string) string {
	rel, _ := url.Parse(name)
	return ps.baseURL.ResolveReference(rel).String()
}

func (ps *PokemonService) getPokemonFromAPI(name string) (*Pokemon, error) {
	url := ps.getPokemonURL(name)
	// url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%s", name)

	resp, err := ps.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w while trying to retrieve pokemon from api url %s: %v", types.ErrGeneric, url, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w while searching for pokemon %s", types.ErrNotFound, name)

	} else if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("%w unexpected status %d from upstream while searching for pokemon %s: %s", types.ErrGeneric, resp.StatusCode, name, string(bodyBytes))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w while reading response body for pokemon %s: %v", types.ErrGeneric, name, err)
	}

	var apiPokemon APIPokemon
	if err := json.Unmarshal(body, &apiPokemon); err != nil {
		return nil, fmt.Errorf("%w while unmarshaling response for pokemon %s: %v", types.ErrGeneric, name, err)
	}

	internal := apiPokemon.toInternal()
	return &internal, nil
}

func determineTranslationType(pkmn *Pokemon) types.Translation {
	if strings.ToLower(pkmn.Habitat) == "cave" || pkmn.IsLegendary {
		return types.Yoda
	}
	return types.Shakespeare
}

func (ps *PokemonService) getPokemon(name string) (*Pokemon, error) {
	cached, exists := ps.cache.Get(name)
	if exists {
		return cached.(*Pokemon), nil
	}

	internal, err := ps.getPokemonFromAPI(name)
	if err != nil {
		return nil, err
	}
	ps.cache.Put(name, internal)
	return internal, nil
}

func (ps *PokemonService) GetPokemon(name string, translate bool) (*Pokemon, error) {
	name = normalize(name)
	pkmn, err := ps.getPokemon(name)
	if err != nil {
		return nil, err
	}

	if !translate || pkmn.Desc == "" {
		return pkmn, nil
	}

	translation := determineTranslationType(pkmn)
	translated, err := ps.translator.Translate(name, pkmn.Desc, translation)
	if err != nil {
		slog.Error("failed to translate description of pokemon %s: %v", name, err)
		return pkmn, nil
	}

	p := *pkmn
	p.Desc = *translated
	return &p, nil
}
