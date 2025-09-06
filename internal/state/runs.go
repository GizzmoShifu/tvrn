package state

import (
  "encoding/json"
  "os"
  "path/filepath"
  "time"
)

type RunRecord struct {
  Time   time.Time `json:"time"`
  Before string    `json:"before"`
  After  string    `json:"after"`
  Error  string    `json:"error,omitempty"`
}

func AppendRun(home string, rec RunRecord) {
  f := filepath.Join(home, "state", "last_run.jsonl")
  fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
  if err != nil { return }
  defer fd.Close()
  b, _ := json.Marshal(rec)
  fd.Write(append(b, '\n'))
}
