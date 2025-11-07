package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/sbaglivi/TL-Pokedex/types"
)

type PokemonService interface {
	GetPokemon(ctx context.Context, name string, translate bool) (*types.GetPokemonResult, error)
}

type Handler struct {
	pkmnSvc PokemonService
}

func NewHandler(pkmnSvc PokemonService) *Handler {
	return &Handler{pkmnSvc: pkmnSvc}
}

func (h *Handler) Register(app *fiber.App) {
	v1 := app.Group("/api/v1")
	v1.Get("/pokemon/:name", timeout.NewWithContext(h.GetPokemon, time.Second*5))
	v1.Get("/pokemon/translated/:name", timeout.NewWithContext(h.GetPokemonWithTranslation, time.Second*9))
}

func handleError(c *fiber.Ctx, err error, logMsg string) error {
	switch {
	case errors.Is(err, context.Canceled):
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return c.Status(fiber.StatusGatewayTimeout).JSON(types.Timeout.Wrap())
	case errors.Is(err, types.ErrNotFound):
		return c.Status(404).JSON(types.NotFound.Wrap())
	default:
		slog.Error(logMsg, "error", err)
		return c.Status(500).JSON(types.InternalServerError.Wrap())
	}
}

func (h *Handler) GetPokemon(c *fiber.Ctx) error {
	name := c.Params("name")
	ctx := c.UserContext()
	pkmn, err := h.pkmnSvc.GetPokemon(ctx, name, false)

	if err != nil {
		return handleError(c, err, "failed to get pokemon")
	}

	return c.Status(200).JSON(pkmn)
}

func (h *Handler) GetPokemonWithTranslation(c *fiber.Ctx) error {
	name := c.Params("name")
	ctx := c.UserContext()
	pkmn, err := h.pkmnSvc.GetPokemon(ctx, name, true)

	if err != nil {
		return handleError(c, err, "failed to get pokemon")
	}

	return c.Status(200).JSON(pkmn)
}
