package config

import (
  "errors"
  "fmt"
  "os"
  "path/filepath"

  "github.com/pelletier/go-toml/v2"
)

type Config struct {
  Home    string    `toml:"-"`
  Auth    Auth      `toml:"auth"`
  Cache   Cache     `toml:"cache"`
  Rename  Rename    `toml:"rename"`
  Defaults Defaults `toml:"defaults"`
  CLI     CLI       `toml:"-"`
  Log     Log       `toml:"log"`
}

type Auth struct {
  APIKey string `toml:"apikey"`
  PIN    string `toml:"pin"`
}

type Cache struct {
  EpisodesTTLHours int  `toml:"episodes_ttl_hours"`
  SeriesTTLDays    int  `toml:"series_ttl_days"`
  SearchTTLDays    int  `toml:"search_ttl_days"`
  ValidateWithETag bool `toml:"validate_with_etag"`
}

type Rename struct {
  Scheme     string `toml:"scheme"`
  Pad        int    `toml:"pad"`
  Specials   string `toml:"specials"`
  MultiEP    string `toml:"multi_ep"`
  DateInName string `toml:"date_in_title"`
  TagsRegex  string `toml:"tags_pattern"`
}

type Defaults struct {
  Order               string `toml:"order"`
  Lang                string `toml:"lang"`
  ConfirmationStrict  bool   `toml:"confirmation_strict"`
}

type CLI struct {
  Root     string
  Scheme   string
  Pad      int
  Order    string
  Lang     string
  Specials string
  MultiEP  string
  Season   int
  Detailed bool
  Debug    bool
  NoCache  bool
  ForceRef bool
  Yes      bool
}

type Log struct { Level string `toml:"level"` }

func Load() (*Config, error) {
  home := os.Getenv("TVRN_HOME")
  if home == "" {
    dir, err := os.UserHomeDir()
    if err != nil { return nil, err }
    home = filepath.Join(dir, ".tvrn")
  }
  if err := os.MkdirAll(filepath.Join(home, "cache"), 0o755); err != nil { return nil, err }
  if err := os.MkdirAll(filepath.Join(home, "state"), 0o755); err != nil { return nil, err }
  if err := os.MkdirAll(filepath.Join(home, "logs"), 0o755); err != nil { return nil, err }

  cfg := &Config{Home: home}
  // sensible defaults
  cfg.Auth = Auth{APIKey: os.Getenv("TVDB_APIKEY"), PIN: os.Getenv("TVDB_PIN")}
  cfg.Cache = Cache{EpisodesTTLHours: 24, SeriesTTLDays: 7, SearchTTLDays: 7, ValidateWithETag: true}
  cfg.Rename = Rename{Scheme: defaultScheme, Pad: defaultPad, Specials: "inline", MultiEP: "range", DateInName: "none"}
  cfg.Defaults = Defaults{Order: defaultOrder, Lang: defaultLang, ConfirmationStrict: true}
  cfg.Log = Log{Level: "info"}

  path := filepath.Join(home, "config.toml")
  if b, err := os.ReadFile(path); err == nil {
    if err := toml.Unmarshal(b, cfg); err != nil {
      return nil, fmt.Errorf("parse config: %w", err)
    }
  }
  if cfg.Auth.APIKey == "" {
    return nil, errors.New("TVDB API key missing: set auth.apikey or TVDB_APIKEY")
  }
  return cfg, nil
}
