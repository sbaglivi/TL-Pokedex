package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type NameAndURL struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type FlavorTextEntry struct {
	FlavorText string     `json:"flavor_text"`
	Language   NameAndURL `json:"language"`
	Version    NameAndURL `json:"version"`
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

var re = regexp.MustCompile(`\s+`)

func removeWhitespace(s string) string {
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
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

func main() {
	name := "pikachu"
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%s", name)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var apiPokemon APIPokemon
	if err := json.Unmarshal(body, &apiPokemon); err != nil {
		panic(err)
	}

	internalPokemon := apiPokemon.toInternal()

	fmt.Printf("Parsed pokemon:\nName: %s\nHabitat: %s\nLegendary: %t\nDescription: %s\n", internalPokemon.Name, internalPokemon.Habitat, internalPokemon.IsLegendary, internalPokemon.Desc)
}
