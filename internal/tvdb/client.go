package tvdb

import "context"

// Client is the surface the runner depends on.
// Concrete implementation (HTTPClient) lives in http.go.
type Client interface {
  Login(ctx context.Context) error
  SearchSeries(ctx context.Context, q, lang string) ([]Series, error)
  GetSeries(ctx context.Context, id int, lang string) (Series, error)
  GetEpisodes(ctx context.Context, id int, order string, season int, lang string) ([]Episode, error)
}
