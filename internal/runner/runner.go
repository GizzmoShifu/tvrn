package runner

import (
  "context"
  "fmt"
  "io"
  "os"
  "path/filepath"
  "regexp"
  "strconv"
  "strings"
  "runtime"

  "github.com/GizzmoShifu/tvrn/internal/config"
  "github.com/GizzmoShifu/tvrn/internal/logx"
  "github.com/GizzmoShifu/tvrn/internal/parse"
  "github.com/GizzmoShifu/tvrn/internal/planner"
  "github.com/GizzmoShifu/tvrn/internal/state"
  "github.com/GizzmoShifu/tvrn/internal/tvdb"
)

type Runner struct {
  cfg  *config.Config
  log  *logx.Logger
  pins *state.Pins
  tv   tvdb.Client
}

func New(cfg *config.Config, log *logx.Logger, tv tvdb.Client) *Runner {
  p, _ := state.LoadPins(cfg.Home)
  return &Runner{cfg: cfg, log: log, pins: p, tv: tv}
}

func (r *Runner) Cfg() *config.Config { return r.cfg }

func (r *Runner) debugf(format string, a ...any) {
  if r.cfg.CLI.Debug {
    fmt.Fprintf(os.Stderr, "DEBUG "+format+"\n", a...)
  }
}

var seasonDirRe = regexp.MustCompile(`(?i)^s(eason)?\s*0?\d+$|^specials$`)

func (r *Runner) Plan(ctx context.Context, root string) (planner.Plan, planner.Stats, error) {
  // Work out series name and season hint from the path
  base := filepath.Base(root)
  parent := filepath.Base(filepath.Dir(root))

  seriesName := base
  seasonHint := 0
  if seasonDirRe.MatchString(base) {
    seriesName = parent
    if digits := regexp.MustCompile(`\d+`).FindString(base); digits != "" {
      fmt.Sscanf(digits, "%d", &seasonHint)
    }
  }

  // Optional year hint e.g. "Firefly (2002)"
  yearHint := 0
  if i := strings.LastIndex(seriesName, "("); i > 0 && strings.HasSuffix(seriesName, ")") {
    if y, err := strconv.Atoi(strings.TrimRight(seriesName[i+1:], ")")); err == nil {
      yearHint = y
      seriesName = strings.TrimSpace(seriesName[:i])
    }
  }

  // TVDB client
  c := r.tv
  if c == nil {
    c = tvdb.NewHTTP("", r.cfg.Auth.APIKey, r.cfg.Auth.PIN)
  }
  if err := c.Login(ctx); err != nil { return planner.Plan{}, planner.Stats{}, err }

  // Search series
  hits, err := c.SearchSeries(ctx, seriesName, r.cfg.Defaults.Lang)
  if err != nil { return planner.Plan{}, planner.Stats{}, err }
  if len(hits) == 0 { return planner.Plan{}, planner.Stats{}, fmt.Errorf("no TVDB results for %q", seriesName) }

  // Pick best match
  show := hits[0]
  for _, h := range hits {
    if strings.EqualFold(h.Name, seriesName) && (yearHint == 0 || h.Year == yearHint) {
      show = h
      break
    }
  }

  // Fetch episodes for configured order and the current season only
  eps, err := c.GetEpisodes(ctx, show.ID, r.cfg.Defaults.Order, seasonHint, r.cfg.Defaults.Lang)
  if err != nil { return planner.Plan{}, planner.Stats{}, err }
  r.debugf("picked series=%q id=%d order=%s season=%d; fetched episodes=%d",
    show.Name, show.ID, r.cfg.Defaults.Order, seasonHint, len(eps))
  for i := 0; i < len(eps) && i < 5; i++ {
    e := eps[i]
    r.debugf("api sample: S%02dE%02d -> %q", e.Season, e.Number, e.Title)
  }

  // Index episodes by S/E
  type key struct{ s, e int }
  bySE := map[key]tvdb.Episode{}
  for _, e := range eps { bySE[key{e.Season, e.Number}] = e }

  // Walk current directory for media files
  entries, err := os.ReadDir(root)
  if err != nil { return planner.Plan{}, planner.Stats{}, err }

  var plan planner.Plan
  skipped := 0
  for _, ent := range entries {
    if ent.IsDir() { continue }
    name := ent.Name()
    lower := strings.ToLower(name)
    if !(strings.HasSuffix(lower, ".mkv") || strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".avi")) {
      continue
    }

    p, ok := parse.FromFilename(name, seasonHint, "")
    if !ok {
      r.debugf("parse miss: %q", name)
      continue
    }

    title := ""
    if e, ok := bySE[key{p.Season, p.Episode}]; ok {
      title = e.Title
    }
    if p.Episode2 > 0 && p.Episode2 > p.Episode {
      if e2, ok := bySE[key{p.Season, p.Episode2}]; ok && e2.Title != "" {
        if title != "" { title = title + " + " + e2.Title } else { title = e2.Title }
      }
    }
    r.debugf("file=%q parsed=S%02dE%02d%s title=%q",
      name, p.Season, p.Episode,
      func() string {
        if p.Episode2 > 0 { return fmt.Sprintf("-%02d", p.Episode2) }
        return ""
      }(),
      title,
    )

    toName := formatName(r.cfg.Rename.Scheme, r.cfg.Rename.Pad, r.cfg.Rename.MultiEP,
      seriesName, p.Season, p.Episode, p.Episode2, title, p.Ext)

    // Skip no-ops where the file is already correctly named
    if sameFileName(name, toName) {
      r.debugf("noop (already named): %q", name)
      skipped++
      continue
    }

    plan.Items = append(plan.Items, planner.Item{
      From:   filepath.Join(root, name),
      To:     filepath.Join(root, toName),
      Reason: "rename",
    })
  }

  st := planner.Stats{Total: len(plan.Items), Skipped: skipped}
  for _, it := range plan.Items {
    if _, err := os.Stat(it.To); err == nil { st.Collisions++ }
  }
  return plan, st, nil
}

func (r *Runner) PrintPreview(p planner.Plan, detailed bool) {
  for _, it := range p.Items {
    if detailed {
      fmt.Printf("%s -> %s\n", filepath.Base(it.From), filepath.Base(it.To))
    } else {
      fmt.Println(filepath.Base(it.To))
    }
  }
}

func (r *Runner) Confirm(in io.Reader, out io.Writer, n int) (bool, error) {
  return confirm(in, out, n, r.cfg.Defaults.ConfirmationStrict)
}

type ApplyResult struct{ Total, Errors int }

func (r *Runner) Apply(ctx context.Context, p planner.Plan) ApplyResult {
  var res ApplyResult
  res.Total = len(p.Items)
  for _, it := range p.Items {
    if _, err := os.Stat(it.To); err == nil {
      r.log.Warnf("skip (exists): %s", it.To)
      continue
    }
    if err := os.Rename(it.From, it.To); err != nil {
      r.log.Errorf("rename failed: %s -> %s: %v", it.From, it.To, err)
      res.Errors++
      continue
    }
  }
  return res
}

func (r *Runner) Report(res ApplyResult) {
  fmt.Printf("Applied %d, errors %d\n", res.Total, res.Errors)
}

// formatName: drop series in filename. e.g. "1x03 - Title.mkv" or "1x01-1x02.mkv"
func formatName(scheme string, pad int, multi string, _show string, season, ep, ep2 int, title, ext string) string {
  if scheme == "" { scheme = "XxYY" }
  if pad <= 0 { pad = 2 }

  epFmt := func(s, e int) string {
    switch scheme {
    case "SXXEYY":
      return fmt.Sprintf("S%02dE%0*d", s, pad, e)
    case "sXXeYY":
      return fmt.Sprintf("s%02de%0*d", s, pad, e)
    case "XYY":
      return fmt.Sprintf("%d%0*d", s, pad, e)
    case "YY":
      return fmt.Sprintf("%0*d", pad, e)
    default:
      return fmt.Sprintf("%dx%0*d", s, pad, e) // XxYY
    }
  }

  secondInRange := func(scheme string, pad, e int) string {
    switch scheme {
    case "SXXEYY":
      return fmt.Sprintf("E%0*d", pad, e)     // S01E01-E02
    case "sXXeYY":
      return fmt.Sprintf("e%0*d", pad, e)     // s01e01-e02
    case "XxYY", "XYY", "YY":
      return fmt.Sprintf("%0*d", pad, e)      // 1x01-02, 101-02, 01-02
    default:
      return fmt.Sprintf("%0*d", pad, e)
    }
  }

  epPart := epFmt(season, ep)
  if ep2 > 0 && ep2 > ep {
    if multi == "" || strings.EqualFold(multi, "range") {
      epPart = epPart + "-" + secondInRange(scheme, pad, ep2)
      cleanTitle := strings.TrimSpace(title)
      cleanTitle = strings.NewReplacer("/", "-", "\\", "-", ":", " -", "*", "", "?", "", "\"", "'", "<", "(", ">", ")", "|", "-", "\n", " ", "\r", " ").Replace(cleanTitle)
      if cleanTitle != "" {
        return fmt.Sprintf("%s - %s.%s", epPart, cleanTitle, ext)
      }
      return fmt.Sprintf("%s.%s", epPart, ext)
    }
    // join mode: 1x01x02 or S01E01E02
    switch scheme {
    case "SXXEYY", "sXXeYY":
      epPart = epPart + fmt.Sprintf("E%0*d", pad, ep2)
    default:
      epPart = epPart + fmt.Sprintf("x%0*d", pad, ep2)
    }
    return fmt.Sprintf("%s.%s", epPart, ext)
  }

  cleanTitle := strings.TrimSpace(title)
  cleanTitle = strings.NewReplacer(
    "/", "-", "\\", "-",
    ":", " -", "*", "",
    "?", "", "\"", "'",
    "<", "(", ">", ")",
    "|", "-", "\n", " ",
    "\r", " ",
  ).Replace(cleanTitle)

  if cleanTitle != "" {
    return fmt.Sprintf("%s - %s.%s", epPart, cleanTitle, ext)
  }
  return fmt.Sprintf("%s.%s", epPart, ext)
}

// sameFileName returns true when the two basenames are the same.
// Windows is case-insensitive; Unix is case-sensitive.
func sameFileName(a, b string) bool {
  if runtime.GOOS == "windows" {
    return strings.EqualFold(a, b)
  }
  return a == b
}