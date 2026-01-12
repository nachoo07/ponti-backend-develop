package project

import (
	"context"

	pgs "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// WordsSuggesterEnginePort es la interfaz del motor externo de sugerencias.
type WordsSuggesterEnginePort interface {
	Suggest(context.Context, string, string, string, int, int) ([]pgs.Suggestion, int64, error)
	Close() error
	Health(context.Context) error
}

// WordsSuggester adapta WordsSuggesterEnginePort al puerto de dominio.
type WordsSuggester struct {
	eng WordsSuggesterEnginePort
}

// NewWordsSuggester recibe el motor externo más la tabla/columna a usar.
func NewWordsSuggester(eng WordsSuggesterEnginePort) *WordsSuggester {
	return &WordsSuggester{
		eng: eng,
	}
}

func (a *WordsSuggester) Suggest(
	ctx context.Context,
	prefix string,
	page, perPage int,
) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	// Buscar sugerencias y cantidad total
	ext, total, err := a.eng.Suggest(ctx, "projects", "name", prefix, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.ListedProject, len(ext))
	for i, s := range ext {
		out[i] = domain.ListedProject{
			ID:   int64(s.ID),
			Name: s.Text,
		}
	}
	return out, total, nil
}

// Close delega el cierre de recursos al motor externo.
func (a *WordsSuggester) Close() error {
	return a.eng.Close()
}

// Health delega el chequeo de salud al motor externo.
func (a *WordsSuggester) Health(ctx context.Context) error {
	return a.eng.Health(ctx)
}
