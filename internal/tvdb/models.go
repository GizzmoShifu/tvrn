package tvdb

import "time"

type Series struct {
  ID      int
  Name    string
  Year    int
  Slug    string
  Aliases []string
}

type Episode struct {
  ID        int
  Season    int
  Number    int
  Absolute  int
  Title     string
  AirDate   time.Time
  IsSpecial bool
}
