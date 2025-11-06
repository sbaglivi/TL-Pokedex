package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sbaglivi/TL-Pokedex/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPokemonService struct {
	mock.Mock
}

func (m *mockPokemonService) GetPokemon(name string, translated bool) (*types.GetPokemonResult, error) {
	args := m.Called(name, translated)
	return args.Get(0).(*types.GetPokemonResult), args.Error(1)
}

func TestGetPokemon(t *testing.T) {
	app := fiber.New()

	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}

	app.Get("/pokemon/:name", h.GetPokemon)

	expected := &types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "Pikachu"}}
	mockSvc.On("GetPokemon", "pikachu", false).Return(expected, nil)

	req := httptest.NewRequest("GET", "/pokemon/pikachu", nil)
	resp, _ := app.Test(req, -1)

	body, _ := io.ReadAll(resp.Body)
	var got types.GetPokemonResult
	json.Unmarshal(body, &got)
	assert.Equal(t, *expected, got)
	assert.Equal(t, 200, resp.StatusCode)
}
func TestGetPokemonTranslationFail(t *testing.T) {
	app := fiber.New()

	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}

	app.Get("/pokemon/:name", h.GetPokemon)

	expected := &types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "Pikachu", Desc: "An electric pokemon"}, Warnings: []string{"translation failed"}}
	mockSvc.On("GetPokemon", "pikachu", false).Return(expected, nil)

	req := httptest.NewRequest("GET", "/pokemon/pikachu", nil)
	resp, _ := app.Test(req, -1)

	body, _ := io.ReadAll(resp.Body)
	var got types.GetPokemonResult
	json.Unmarshal(body, &got)
	assert.Equal(t, *expected, got)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetPokemon_NotFound(t *testing.T) {
	app := fiber.New()
	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}

	app.Get("/pokemon/:name", h.GetPokemon)

	mockSvc.On("GetPokemon", "missing", false).Return(&types.GetPokemonResult{}, types.ErrNotFound)

	req := httptest.NewRequest("GET", "/pokemon/missing", nil)
	resp, _ := app.Test(req, -1)
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, string(body), "\"not found\"")
}

func TestGetPokemon_InternalError(t *testing.T) {
	app := fiber.New()
	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}

	app.Get("/pokemon/:name", h.GetPokemon)

	mockSvc.On("GetPokemon", "pikachu", false).Return(&types.GetPokemonResult{}, errors.New("db failure"))

	req := httptest.NewRequest("GET", "/pokemon/pikachu", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, 500, resp.StatusCode)
}
