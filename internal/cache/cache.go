package cache

import "time"

type Entry struct {
  Body      []byte
  ETag      string
  Modified  time.Time
  Expires   time.Time
}

type Store interface {
  Get(key string) (Entry, bool)
  Put(key string, e Entry) error
}
