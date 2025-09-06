package parse

import "regexp"

var (
  reSxxExx = regexp.MustCompile(`(?i)S(\d{1,2})E(\d{1,2})(?:[\-E](\d{1,2}))?`)
  reXxYY   = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})(?:[\-x](\d{1,2}))?`)
  reNNN    = regexp.MustCompile(`(?i)(\d)(\d{2})(?:-(\d{2}))?`) // needs season context
)
