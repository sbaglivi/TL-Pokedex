package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sbaglivi/TL-Pokedex/cache"
	"github.com/sbaglivi/TL-Pokedex/handler"
	"github.com/sbaglivi/TL-Pokedex/pokemon"
	"github.com/sbaglivi/TL-Pokedex/translate"
)

func createPokemonService() *pokemon.PokemonService {
	cache := cache.NewLRU(1024)
	client := http.DefaultClient
	translateService, err := translate.NewTranslationService(cache, "https://api.funtranslations.com/translate/", client)

	if err != nil {
		slog.Error("failed to initialize translation service", "error", err)
		os.Exit(1)
	}

	pkmnService, err := pokemon.NewPokemonService(cache, translateService, "https://pokeapi.co/api/v2/pokemon-species/", client)
	if err != nil {
		slog.Error("failed to initialize pokemon service", "error", err)
		os.Exit(1)
	}

	return pkmnService
}

func getPort() int {
	defaultPort := 3000
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return defaultPort
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		slog.Error("failed to parse PORT env var as int", "port", portStr, "error", err)
		os.Exit(1)
	}

	return port
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	pkmnService := createPokemonService()
	handler := handler.NewHandler(pkmnService)

	app := fiber.New()
	app.Get("/pokemon/:name", handler.GetPokemon)
	app.Get("/pokemon/translated/:name", handler.GetPokemonWithTranslation)
	port := getPort()
	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to start server", "port", port, "error", err)
		os.Exit(1)
	}
}
