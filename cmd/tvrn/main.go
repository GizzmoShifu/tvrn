package main

import (
  "context"
  "fmt"
  "os"
  "path/filepath"
  "time"

  "github.com/GizzmoShifu/tvrn/internal/cache"
  "github.com/GizzmoShifu/tvrn/internal/config"
  "github.com/GizzmoShifu/tvrn/internal/logx"
  "github.com/GizzmoShifu/tvrn/internal/runner"
  "github.com/GizzmoShifu/tvrn/internal/tvdb"
)

func main() {
  ctx := context.Background()
  cfg, err := config.Load()
  if err != nil { fatal(err) }

  root := "."
  if cfg.CLI.Root != "" { root = cfg.CLI.Root }
  root, _ = filepath.Abs(root)

  log := logx.New(cfg.Log.Level)
  log.Infof("tvrn starting in %s", root)

  fs := cache.NewFS(cfg.Home)
  client := tvdb.NewHTTP("", cfg.Auth.APIKey, cfg.Auth.PIN).WithCache(fs)
  rn := runner.New(cfg, log, client)

  plan, stats, err := rn.Plan(ctx, root)
  if err != nil { fatal(err) }

  if stats.Total == 0 {
    fmt.Println("No changes needed")
    os.Exit(0)
  }

  rn.PrintPreview(plan, cfg.CLI.Detailed)

  proceed := cfg.CLI.Yes
  if !proceed {
    proceed, err = rn.Confirm(os.Stdin, os.Stdout, stats.Total)
    if err != nil { fatal(err) }
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
}

func fatal(err error) {
  fmt.Fprintf(os.Stderr, "error: %v\n", err)
  time.Sleep(10 * time.Millisecond)
  os.Exit(4)
}