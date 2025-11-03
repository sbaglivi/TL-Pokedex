package main

import (
	"reflect"
	"testing"
)

func TestAPIPokemonToInternal(t *testing.T) {
	apiPokemon := APIPokemon{
		IsLegendary: true,
		Name:        "Groudon",
		APIHabitat: NameAndURL{
			Name: "Lava",
			Url:  "http://",
		},
		FlavorTextEntries: []FlavorTextEntry{
			{
				FlavorText: "   Test\r\nof a weird     description",
				Language: NameAndURL{
					Name: "en",
					Url:  "http://",
				},
				Version: NameAndURL{
					Name: "Ruby",
					Url:  "http://",
				},
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
