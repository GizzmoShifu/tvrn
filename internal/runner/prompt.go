package runner

import (
  "bufio"
  "fmt"
  "io"
  "strings"
)

func confirm(r io.Reader, w io.Writer, n int, strict bool) (bool, error) {
  fmt.Fprintf(w, "Apply %d changes? Type Y to continue: ", n)
  s := bufio.NewScanner(r)
  if !s.Scan() { return false, s.Err() }
  ans := strings.TrimSpace(s.Text())
  if strict { return ans == "Y", nil }
  return strings.EqualFold(ans, "y"), nil
}
