package pokemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/sbaglivi/TL-Pokedex/types"
	"github.com/sbaglivi/TL-Pokedex/utils"
)

type NameAndURL struct {
	Name string `json:"name"`
}

type FlavorTextEntry struct {
	FlavorText string     `json:"flavor_text"`
	Language   NameAndURL `json:"language"`
}

type APIPokemon struct {
	IsLegendary       bool              `json:"is_legendary"`
	Name              string            `json:"name"`
	APIHabitat        NameAndURL        `json:"habitat"`
	FlavorTextEntries []FlavorTextEntry `json:"flavor_text_entries"`
}

type Translator interface {
	Translate(context.Context, string, string, types.Translation) (*string, error)
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

func normalize(s string) string {
	return utils.RemoveWhitespace(strings.ToLower(s))
}

func getDescription(entries *[]FlavorTextEntry) string {
	if len(*entries) == 0 {
		return ""
	}

	preferredLanguage := "en"
	for _, en := range *entries {
		if en.Language.Name == preferredLanguage {
			return en.FlavorText
		}
	}

	return (*entries)[0].FlavorText
}

func (pkmn *APIPokemon) toInternal() types.Pokemon {
	return types.Pokemon{
		IsLegendary: pkmn.IsLegendary,
		Name:        pkmn.Name,
		Habitat:     pkmn.APIHabitat.Name,
		Desc:        utils.RemoveWhitespace(getDescription(&pkmn.FlavorTextEntries)),
	}
}

func (ps *PokemonService) getPokemonURL(name string) string {
	rel, _ := url.Parse(name)
	return ps.baseURL.ResolveReference(rel).String()
}

func (ps *PokemonService) getPokemonFromAPI(ctx context.Context, name string) (*types.Pokemon, error) {
	url := ps.getPokemonURL(name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w while creating req to retrieve pokemon from api with url %s: %v", types.ErrGeneric, url, err)
	}
	resp, err := ps.client.Do(req)
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

func determineTranslationType(pkmn *types.Pokemon) types.Translation {
	if strings.ToLower(pkmn.Habitat) == "cave" || pkmn.IsLegendary {
		return types.Yoda
	}
	return types.Shakespeare
}

func (ps *PokemonService) getPokemon(ctx context.Context, name string) (*types.Pokemon, error) {
	cached, exists := ps.cache.Get(name)
	if exists {
		return cached.(*types.Pokemon), nil
	}

	internal, err := ps.getPokemonFromAPI(ctx, name)
	if err != nil {
		return nil, err
	}
	ps.cache.Put(name, internal)
	return internal, nil
}

func (ps *PokemonService) GetPokemon(ctx context.Context, name string, translate bool) (*types.GetPokemonResult, error) {
	name = normalize(name)
	pkmn, err := ps.getPokemon(ctx, name)
	if err != nil {
		return nil, err
	}

	if !translate || pkmn.Desc == "" {
		return &types.GetPokemonResult{Pokemon: pkmn, Warnings: nil}, nil
	}

	translation := determineTranslationType(pkmn)
	translated, err := ps.translator.Translate(ctx, name, pkmn.Desc, translation)
	if err != nil {
		if !errors.Is(types.ErrTooManyRequests, err) {
			slog.Error("failed to translate description", "pokemon", name, "error", err)
		}
		return &types.GetPokemonResult{Pokemon: pkmn, Warnings: []string{"translation failed"}}, nil
	}

	p := *pkmn
	p.Desc = *translated
	return &types.GetPokemonResult{Pokemon: &p, Warnings: nil}, nil
}
