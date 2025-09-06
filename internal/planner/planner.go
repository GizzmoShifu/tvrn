package planner

import (
  "fmt"
  "regexp"
  "strings"
)

// formatEpisodeName builds the target filename (without any directory path)
// Examples (Pad=2):
//  Scheme=XxYY, single  -> 1x03 - Our Mrs. Reynolds.mkv
//  Scheme=XxYY, range   -> 1x01-1x02.mkv
//  Scheme=SXXEYY, single-> S01E03 - Our Mrs. Reynolds.mkv
//  Scheme=SXXEYY, range -> S01E01-S01E02.mkv
func formatEpisodeName(opts Options, season, ep, ep2 int, title, ext string) string {
  pad := opts.Pad
  if pad < 2 { pad = 2 }

  // number part
  num := func(s, e int) string {
    switch opts.Scheme {
    case "SXXEYY", "sXXeYY":
      prefix := "S"
      if opts.Scheme == "sXXeYY" { prefix = "s" }
      return fmt.Sprintf("%s%0*dE%0*d", prefix, 2, season, pad, e)
    case "XYY":
      return fmt.Sprintf("%d%0*d", season, pad, e)
    case "YY":
      return fmt.Sprintf("%0*d", pad, e)
    case "XxYY", "xXyY", "xxyy": // treat all as 1x01 style
      return fmt.Sprintf("%dx%0*d", season, pad, e)
    default: // fallback to XxYY
      return fmt.Sprintf("%dx%0*d", season, pad, e)
    }
  }

  base := num(season, ep)
  if ep2 > 0 {
    // multi-episode handling
    if strings.EqualFold(opts.MultiEP, "join") {
      // e.g. S01E01E02 or 1x01x02
      // reuse the left part and tack on the second episode number only
      switch opts.Scheme {
      case "SXXEYY", "sXXeYY":
        base = base + fmt.Sprintf("E%0*d", pad, ep2)
      case "XxYY", "xXyY", "xxyy":
        base = base + fmt.Sprintf("x%0*d", pad, ep2)
      default:
        base = base + fmt.Sprintf("-%s", num(season, ep2))
      }
    } else {
      // default: range (numbers only)
      base = base + "-" + num(season, ep2)
    }
    return base + "." + ext
  }

  // single episode: append title if we have one
  t := sanitizeTitle(title)
  if t != "" {
    return fmt.Sprintf("%s - %s.%s", base, t, ext)
  }
  return base + "." + ext
}

var bad = regexp.MustCompile(`[\\/:*?"<>|]+`)
func sanitizeTitle(s string) string {
  s = strings.TrimSpace(s)
  s = bad.ReplaceAllString(s, " ")
  s = strings.Join(strings.Fields(s), " ")
  return s
}
