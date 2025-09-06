package state

import (
  "encoding/json"
  "os"
  "path/filepath"
)

type Pin struct {
  Path    string `json:"path"`
  SeriesID int    `json:"seriesId"`
  Order   string `json:"order"`
  Lang    string `json:"lang"`
  Locked  bool   `json:"locked"`
}

type Pins struct {
  file string
  byPath map[string]Pin
}

func LoadPins(home string) (*Pins, error) {
  p := &Pins{file: filepath.Join(home, "state", "pins.json"), byPath: map[string]Pin{}}
  b, err := os.ReadFile(p.file)
  if err == nil {
    _ = json.Unmarshal(b, &p.byPath)
  }
  return p, nil
}

func (p *Pins) Get(path string) (Pin, bool) { v, ok := p.byPath[path]; return v, ok }

func (p *Pins) Put(pin Pin) error {
  p.byPath[pin.Path] = pin
  b, _ := json.MarshalIndent(p.byPath, "", "  ")
  return os.WriteFile(p.file, b, 0o644)
}
