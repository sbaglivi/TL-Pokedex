package handler

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/sbaglivi/TL-Pokedex/types"
)

type PokemonService interface {
	GetPokemon(name string, translate bool) (*types.GetPokemonResult, error)
}

type Handler struct {
	pkmnSvc PokemonService
}

func NewHandler(pkmnSvc PokemonService) *Handler {
	return &Handler{pkmnSvc: pkmnSvc}
}

func (h *Handler) GetPokemon(c *fiber.Ctx) error {
	name := c.Params("name")
	pkmn, err := h.pkmnSvc.GetPokemon(name, false)

	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return c.Status(404).JSON(types.NotFound.Wrap())
		}

		slog.Error("failed to get pokemon", "error", err)
		return c.Status(500).JSON(types.InternalServerError.Wrap())
	}

	return c.Status(200).JSON(pkmn)
}

func (h *Handler) GetPokemonWithTranslation(c *fiber.Ctx) error {
	name := c.Params("name")
	pkmn, err := h.pkmnSvc.GetPokemon(name, true)

	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return c.Status(404).JSON(types.NotFound.Wrap())
		}

		slog.Error("failed to get pokemon", "error", err)
		return c.Status(500).JSON(types.InternalServerError.Wrap())
	}

	return c.Status(200).JSON(pkmn)
}
