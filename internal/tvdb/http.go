package tvdb

import (
  "bytes"
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "net/http"
  "net/url"
  "path"
  "strconv"
  "strings"
  "time"

  "github.com/GizzmoShifu/tvrn/internal/cache"
)

type HTTPClient struct {
  BaseURL string
  APIKey  string
  PIN     string

  hc       *http.Client
  token    string
  tokenExp time.Time

  cache    cache.Store
}

func NewHTTP(base, apikey, pin string) *HTTPClient {
  if base == "" { base = "https://api4.thetvdb.com/v4" }
  return &HTTPClient{
    BaseURL: base,
    APIKey:  apikey,
    PIN:     pin,
    hc:      &http.Client{Timeout: 20 * time.Second},
  }
}

func (c *HTTPClient) WithCache(s cache.Store) *HTTPClient { c.cache = s; return c }

// ===== Interface methods =====

func (c *HTTPClient) Login(ctx context.Context) error {
  if c.APIKey == "" { return errors.New("missing API key") }
  var lr struct {
    Status string `json:"status"`
    Data   struct{ Token string `json:"token"` } `json:"data"`
  }
  err := c.doJSON(ctx, http.MethodPost, c.u("/login"),
    struct {
      APIKey string `json:"apikey"`
      PIN    string `json:"pin,omitempty"`
    }{APIKey: c.APIKey, PIN: c.PIN},
    "", &lr, false,
  )
  if err != nil { return err }
  if lr.Data.Token == "" { return errors.New("empty token from login") }
  c.token = lr.Data.Token
  c.tokenExp = time.Now().Add(30 * 24 * time.Hour)
  return nil
}

func (c *HTTPClient) SearchSeries(ctx context.Context, q, lang string) ([]Series, error) {
  if err := c.ensureAuth(ctx); err != nil { return nil, err }
  v := url.Values{}
  v.Set("q", q)
  v.Set("type", "series")

  var sr struct {
    Status string `json:"status"`
    Data   []struct {
      ID      any     `json:"tvdb_id"`
      Name    string  `json:"name"`
      Year    any     `json:"year"`
      Slug    string  `json:"slug"`
      Aliases []string`json:"aliases"`
      Type    string  `json:"type"`
    } `json:"data"`
  }
  if err := c.doJSON(ctx, http.MethodGet, c.u("/search")+"?"+v.Encode(), nil, lang, &sr, true); err != nil {
    return nil, err
  }

  out := make([]Series, 0, len(sr.Data))
  for _, d := range sr.Data {
    if strings.ToLower(d.Type) != "series" { continue }
    out = append(out, Series{ID: intFromAny(d.ID), Name: d.Name, Year: intFromAny(d.Year), Slug: d.Slug, Aliases: d.Aliases})
  }
  return out, nil
}

func (c *HTTPClient) GetSeries(ctx context.Context, id int, lang string) (Series, error) {
  if err := c.ensureAuth(ctx); err != nil { return Series{}, err }

  // tolerate string-or-number fields (id/year)
  var sr struct {
    Status string `json:"status"`
    Data   struct {
      ID      any      `json:"id"`
      Name    string   `json:"name"`
      Slug    string   `json:"slug"`
      Year    any      `json:"year"`
      Aliases []string `json:"aliases"`
    } `json:"data"`
  }

  if err := c.doJSON(
    ctx,
    http.MethodGet,
    c.u(path.Join("/series", strconv.Itoa(id))),
    nil,           // no body
    lang,          // Accept-Language
    &sr,
    true,          // with auth
  ); err != nil {
    return Series{}, err
  }

  d := sr.Data
  return Series{
    ID:      intFromAny(d.ID),
    Name:    d.Name,
    Slug:    d.Slug,
    Year:    intFromAny(d.Year),
    Aliases: d.Aliases,
  }, nil
}

func (c *HTTPClient) GetEpisodes(ctx context.Context, id int, order string, season int, _lang string) ([]Episode, error) {
  if err := c.ensureAuth(ctx); err != nil { return nil, err }
  order = normaliseOrder(order)

  type episodesResp struct {
    Status string `json:"status"`
    Data struct {
      Episodes []struct {
        ID       any    `json:"id"`
        Name     string `json:"name"`
        Aired    string `json:"aired"`
        Number   any    `json:"number"`
        Absolute any    `json:"absoluteNumber"`
        Season   any    `json:"seasonNumber"`
      } `json:"episodes"`
    } `json:"data"`
    Links struct{ Next any `json:"next"` } `json:"links"`
  }

  page := 0
  out := make([]Episode, 0, 64)
  for {
    q := url.Values{}
    q.Set("page", strconv.Itoa(page))
    if season > 0 { q.Set("season", strconv.Itoa(season)) }

    var er episodesResp
    if err := c.doJSON(ctx, http.MethodGet,
      c.u(path.Join("/series", strconv.Itoa(id), "episodes", order))+"?"+q.Encode(),
      nil, "", &er, true); err != nil {
      return nil, err
    }

    for _, d := range er.Data.Episodes {
      var t time.Time
      if d.Aired != "" {
        if tt, _ := time.Parse("2006-01-02", d.Aired); !tt.IsZero() { t = tt }
      }
      out = append(out, Episode{
        ID:       intFromAny(d.ID),
        Title:    d.Name,
        AirDate:  t,
        Season:   intFromAny(d.Season),
        Number:   intFromAny(d.Number),
        Absolute: intFromAny(d.Absolute),
      })
    }

    next := 0
    switch v := er.Links.Next.(type) {
    case float64: next = int(v)
    case string:  if n, _ := strconv.Atoi(v); n > 0 { next = n }
    }
    if next == 0 { break }
    page = next
  }
  return out, nil
}

// ===== helpers =====

func (c *HTTPClient) ensureAuth(ctx context.Context) error {
  if c.token == "" || time.Now().After(c.tokenExp.Add(-2*time.Minute)) {
    return c.Login(ctx)
  }
  return nil
}

func cacheKeyEpisodes(id int, order string, season int, lang string) string {
  return fmt.Sprintf("episodes:%d:%s:%d:%s", id, normaliseOrder(order), season, strings.ToLower(lang))
}

func intFromAny(v any) int {
  switch t := v.(type) {
  case float64: return int(t)
  case string:
    n, _ := strconv.Atoi(strings.TrimSpace(t))
    return n
  case json.Number:
    n, _ := t.Int64()
    return int(n)
  default:
    return 0
  }
}

func normaliseOrder(order string) string {
  switch strings.ToLower(strings.TrimSpace(order)) {
  case "", "aired", "default": return "default"
  case "dvd": return "dvd"
  case "absolute", "abs": return "absolute"
  case "alternate", "alt": return "alternate"
  case "regional": return "regional"
  case "alternate-dvd", "alternate_dvd", "altdvd": return "alternate-dvd"
  default: return order
  }
}

func (c *HTTPClient) u(p string) string {
  return strings.TrimRight(c.BaseURL, "/") + p
}

const userAgent = "tvrn/0.1 (+https://github.com/GizzmoShifu/tvrn)"

func retryAfterDelay(h string) time.Duration {
  h = strings.TrimSpace(h)
  if h == "" { return 2 * time.Second }
  if n, err := strconv.Atoi(h); err == nil && n > 0 { return time.Duration(n) * time.Second }
  if t, err := http.ParseTime(h); err == nil {
    if d := time.Until(t); d > 0 { return d }
  }
  return 2 * time.Second
}

// doJSON sends the request, retries on 429, optionally re-logins on 401, and decodes JSON into out
func (c *HTTPClient) doJSON(ctx context.Context, method, urlStr string, body any, acceptLang string, out any, withAuth bool) error {
  // marshal once so we can reuse on retries
  var payload []byte
  if body != nil {
    b, err := json.Marshal(body)
    if err != nil { return err }
    payload = b
  }

  for attempt := 0; attempt < 3; attempt++ {
    var rdr io.Reader
    if payload != nil { rdr = bytes.NewReader(payload) }

    req, err := http.NewRequestWithContext(ctx, method, urlStr, rdr)
    if err != nil { return err }
    req.Header.Set("User-Agent", userAgent)
    if acceptLang != "" { req.Header.Set("Accept-Language", acceptLang) }
    if payload != nil { req.Header.Set("Content-Type", "application/json") }
    if withAuth && c.token != "" { req.Header.Set("Authorization", "Bearer "+c.token) }

    resp, err := c.hc.Do(req)
    if err != nil { return err }

    // 429 backoff & retry
    if resp.StatusCode == http.StatusTooManyRequests {
      d := retryAfterDelay(resp.Header.Get("Retry-After"))
      resp.Body.Close()
      time.Sleep(d)
      continue
    }

    // one re-login on 401 when auth was requested
    if resp.StatusCode == http.StatusUnauthorized && withAuth && attempt == 0 {
      resp.Body.Close()
      if err := c.Login(ctx); err != nil { return err }
      continue
    }

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
      b, _ := io.ReadAll(resp.Body)
      resp.Body.Close()
      return fmt.Errorf("%s %s failed: %s: %s", method, urlStr, resp.Status, string(b))
    }

    decErr := json.NewDecoder(resp.Body).Decode(out)
    resp.Body.Close()
    return decErr
  }
  return fmt.Errorf("%s %s failed after retries", method, urlStr)
}
