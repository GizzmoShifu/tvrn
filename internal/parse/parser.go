package parse

import (
  "path/filepath"
  "strconv"
  "strings"
)

type Parsed struct {
  Show     string
  Season   int
  Episode  int
  Episode2 int // end of range; 0 means single
  Ext      string
  Raw      string
}

func atoi(s string) int { i, _ := strconv.Atoi(s); return i }

func FromFilename(name string, seasonHint int, showHint string) (Parsed, bool) {
  p := Parsed{Raw: name, Ext: strings.TrimPrefix(filepath.Ext(name), ".")}
  base := strings.TrimSuffix(name, filepath.Ext(name))
  s := base

  if m := reSxxExx.FindStringSubmatch(s); len(m) > 0 {
    p.Season = atoi(m[1])
    p.Episode = atoi(m[2])
    if len(m) > 3 && m[3] != "" { p.Episode2 = atoi(m[3]) }
  } else if m := reXxYY.FindStringSubmatch(s); len(m) > 0 {
    p.Season = atoi(m[1])
    p.Episode = atoi(m[2])
    if len(m) > 3 && m[3] != "" { p.Episode2 = atoi(m[3]) }
  } else if m := reNNN.FindStringSubmatch(s); len(m) > 0 && seasonHint > 0 {
    p.Season = seasonHint
    p.Episode = atoi(m[2])
    if len(m) > 3 && m[3] != "" { p.Episode2 = atoi(m[3]) }
  } else {
    return Parsed{}, false
  }

  // Heuristic for show name: prefer hint, else folder name segments before match
  if showHint != "" {
    p.Show = showHint
  } else {
    // crude: replace separators with spaces and trim
    cleaned := strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(s)
    p.Show = strings.TrimSpace(cleaned)
  }
  return p, true
}
