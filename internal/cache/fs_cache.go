package cache

import (
  "encoding/json"
  "os"
  "path/filepath"
  "time"
)

type FS struct { dir string }

func NewFS(home string) *FS { return &FS{dir: filepath.Join(home, "cache")} }

func (f *FS) path(k string) string { return filepath.Join(f.dir, k+".json") }

func (f *FS) Get(k string) (Entry, bool) {
  var e Entry
  b, err := os.ReadFile(f.path(k))
  if err != nil { return e, false }
  if json.Unmarshal(b, &e) != nil { return Entry{}, false }
  if !e.Expires.IsZero() && time.Now().After(e.Expires) { return Entry{}, false }
  return e, true
}

func (f *FS) Put(k string, e Entry) error {
  if err := os.MkdirAll(f.dir, 0o755); err != nil { return err }
  b, _ := json.MarshalIndent(e, "", "  ")
  return os.WriteFile(f.path(k), b, 0o644)
}
