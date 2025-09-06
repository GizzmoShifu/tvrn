package runner

import (
  "bufio"
  "fmt"
  "io"
  "strings"
)

func confirm(r io.Reader, w io.Writer, n int) (bool, error) {
  fmt.Fprintf(w, "\nApply %d changes? Type \"Y\" or \"y\" or \"yes\" to continue: ", n)
  s := bufio.NewScanner(r)
  if !s.Scan() { return false, s.Err() }
  ans := strings.ToLower(strings.TrimSpace(s.Text()))
  return ans == "y" || ans == "yes", nil
}
