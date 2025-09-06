package main

import (
  "context"
  "flag"
  "fmt"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/GizzmoShifu/tvrn/internal/cache"
  "github.com/GizzmoShifu/tvrn/internal/config"
  "github.com/GizzmoShifu/tvrn/internal/logx"
  "github.com/GizzmoShifu/tvrn/internal/runner"
  "github.com/GizzmoShifu/tvrn/internal/tvdb"
)

func main() {
  cfg, err := config.Load()
  if err != nil { fatal(err) }

  // Flags
  fs := flag.NewFlagSet("tvrn", flag.ContinueOnError)
  fs.SetOutput(os.Stdout)

  root := fs.String("root", "", "Root directory to operate on (defaults to current directory)")
  scheme := fs.String("scheme", "", "Episode number format: SXXEYY | sXXeYY | XxYY | XYY | YY")
  pad := fs.Int("pad", 0, "Pad episode number to N digits (default 2)")
  order := fs.String("order", "", "Episode order: aired | dvd | absolute | alternate | regional")
  lang := fs.String("lang", "", "Language code for titles, e.g. en")
  multi := fs.String("multi", "", "Multi-episode naming: range | join")
  season := fs.Int("season", 0, "Force season number when parsing")
  detailed := fs.Bool("detailed", false, "Show before -> after in the proposal")
  debug := fs.Bool("debug", false, "Enable debug logging and verbose matching output")
  seriesMode := fs.Bool("series", false, "Run from a series root and process all season subfolders")
  noCache := fs.Bool("no-cache", false, "Ignore local API cache for this run")
  yes := fs.Bool("yes", false, "Auto-confirm (non-interactive)")

  fs.Usage = func() {
    fmt.Fprintf(os.Stdout, "tvrn – TV renamer using TVDB v4\n\nUsage:\n  tvrn [options] [path]\n\nOptions:\n")
    fs.PrintDefaults()
    fmt.Fprintln(os.Stdout, `
Examples:
  # In a season directory
  tvrn

  # From a series root, process all "Season *" dirs
  tvrn --series

  # Change scheme and pad
  tvrn --scheme=SXXEYY --pad=3

  # Use DVD order and show before->after
  tvrn --order=dvd --detailed`)
  }

  // Allow positional [path]
  if err := fs.Parse(os.Args[1:]); err != nil {
    if err == flag.ErrHelp { os.Exit(0) }
    fatal(err)
  }
  pathArg := "."
  if fs.NArg() > 0 {
    pathArg = fs.Arg(0)
  }
  if *root != "" {
    pathArg = *root
  }
  absRoot, _ := filepath.Abs(pathArg)

  // Merge flags into config
  if *scheme != "" { cfg.Rename.Scheme = *scheme }
  if *pad > 0 { cfg.Rename.Pad = *pad }
  if *order != "" { cfg.Defaults.Order = strings.ToLower(*order) }
  if *lang != "" { cfg.Defaults.Lang = *lang }
  if *multi != "" { cfg.Rename.MultiEP = strings.ToLower(*multi) }
  cfg.CLI.Season = *season
  cfg.CLI.Detailed = *detailed
  cfg.CLI.Debug = *debug
  cfg.CLI.NoCache = *noCache
  cfg.CLI.Yes = *yes
  cfg.CLI.Root = absRoot

  // Logging
  level := "info"
  if cfg.CLI.Debug { level = "debug" }
  log := logx.New(level)
  log.Infof("tvrn starting in %s", absRoot)

  // TVDB client with optional cache
  var client tvdb.Client
  httpc := tvdb.NewHTTP("", cfg.Auth.APIKey, cfg.Auth.PIN)
  if !cfg.CLI.NoCache {
    httpc = httpc.WithCache(cache.NewFS(cfg.Home))
  }
  client = httpc

  rn := runner.New(cfg, log, client)

  // Series mode: discover season subfolders and process them serially
  if *seriesMode {
    // Simple heuristic: dirs named "Season *" or "Specials"
    entries, err := os.ReadDir(absRoot)
    if err != nil { fatal(err) }
    total := 0
    for _, e := range entries {
      if !e.IsDir() { continue }
      name := e.Name()
      lower := strings.ToLower(name)
      if strings.HasPrefix(lower, "season ") || lower == "specials" {
        ok := runOnce(rn, filepath.Join(absRoot, name))
        if ok { total++ }
      }
    }
    if total == 0 { fmt.Println("No season folders found") }
    return
  }

  // Single directory
  runOnce(rn, absRoot)
}

func runOnce(rn *runner.Runner, dir string) bool {
  ctx := context.Background()
  plan, stats, err := rn.Plan(ctx, dir)
  if err != nil { fatal(err) }

  if stats.Total == 0 {
    fmt.Println("No changes needed")
    return false
  }

  rn.PrintPreview(plan, rn.Cfg().CLI.Detailed)

  proceed := rn.Cfg().CLI.Yes
  if !proceed {
    var cerr error
    proceed, cerr = rn.Confirm(os.Stdin, os.Stdout, stats.Total)
    if cerr != nil { fatal(cerr) }
  }
  if !proceed {
    fmt.Println("Cancelled")
    os.Exit(3)
  }

  res := rn.Apply(ctx, plan)
  rn.Report(res)

  if res.Errors > 0 && res.Errors < res.Total {
    os.Exit(2)
  }
  if res.Errors > 0 {
    os.Exit(4)
  }
  return true
}

func fatal(err error) {
  fmt.Fprintf(os.Stderr, "error: %v\n", err)
  time.Sleep(10 * time.Millisecond)
  os.Exit(4)
}
