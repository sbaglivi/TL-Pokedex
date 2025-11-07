package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sbaglivi/TL-Pokedex/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPokemonService struct {
	mock.Mock
}

func (m *mockPokemonService) GetPokemon(ctx context.Context, name string, translated bool) (*types.GetPokemonResult, error) {
	args := m.Called(ctx, name, translated)
	return args.Get(0).(*types.GetPokemonResult), args.Error(1)
}

func TestGetPokemon(t *testing.T) {
	app := fiber.New()

	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}

	h.Register(app)

	expected := &types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "Pikachu"}}
	mockSvc.On("GetPokemon", mock.Anything, "pikachu", false).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/v1/pokemon/pikachu", nil)
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
	h.Register(app)

	expected := &types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "Pikachu", Desc: "An electric pokemon"}, Warnings: []string{"translation failed"}}
	mockSvc.On("GetPokemon", mock.Anything, "pikachu", false).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/v1/pokemon/pikachu", nil)
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
	h.Register(app)

	mockSvc.On("GetPokemon", mock.Anything, "missing", false).Return(&types.GetPokemonResult{}, types.ErrNotFound)

	req := httptest.NewRequest("GET", "/api/v1/pokemon/missing", nil)
	resp, _ := app.Test(req, -1)
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, string(body), "{\"error\":\"not found\"}")
}

func TestGetPokemon_InternalError(t *testing.T) {
	app := fiber.New()
	mockSvc := new(mockPokemonService)
	h := &Handler{pkmnSvc: mockSvc}
	h.Register(app)

	mockSvc.On("GetPokemon", mock.Anything, "pikachu", false).Return(&types.GetPokemonResult{}, errors.New("db failure"))

	req := httptest.NewRequest("GET", "/api/v1/pokemon/pikachu", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, 500, resp.StatusCode)
}

type slowMockPokemonService struct {
	mock.Mock
}

func (m *slowMockPokemonService) GetPokemon(ctx context.Context, name string, translated bool) (*types.GetPokemonResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(2 * time.Second):
		return &types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "pikachu"}}, nil
	}
}

func TestGetPokemonTimeout(t *testing.T) {

	app := fiber.New()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	app.Use(func(c *fiber.Ctx) error {
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	})

	svc := new(slowMockPokemonService)
	svc.On("GetPokemon", mock.Anything, "pikachu", false).
		Return(&types.GetPokemonResult{Pokemon: &types.Pokemon{Name: "pikachu"}}, nil)

	h := NewHandler(svc)
	h.Register(app)

	req := httptest.NewRequest("GET", "/api/v1/pokemon/pikachu", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, fiber.StatusGatewayTimeout, resp.StatusCode)
}
