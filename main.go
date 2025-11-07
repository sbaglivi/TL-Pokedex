package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/sbaglivi/TL-Pokedex/cache"
	"github.com/sbaglivi/TL-Pokedex/handler"
	"github.com/sbaglivi/TL-Pokedex/pokemon"
	"github.com/sbaglivi/TL-Pokedex/translate"
	"github.com/sbaglivi/TL-Pokedex/utils"
)

func createPokemonService() (*pokemon.PokemonService, error) {
	cache := cache.NewLRU(1024)
	client := http.DefaultClient
	translateService, err := translate.NewTranslationService(cache, "https://api.funtranslations.com/translate/", client)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize translation service: %w", err)
	}

	pkmnService, err := pokemon.NewPokemonService(cache, translateService, "https://pokeapi.co/api/v2/pokemon-species/", client)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pokemon service: %w", err)
	}

	return pkmnService, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	pkmnService, err := createPokemonService()
	if err != nil {
		slog.Error("during createPokemonService", "error", err)
		os.Exit(1)
	}

	app := fiber.New()
	handler := handler.NewHandler(pkmnService)
	handler.Register(app)
	port, err := utils.GetPort()
	if err != nil {
		slog.Error("failed to parse PORT env var", "error", err)
		os.Exit(1)
	}

	err = app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to start server", "port", port, "error", err)
		os.Exit(1)
	}
}
