# tvrn – TV renamer (TVDB v4)

Opinionated, safe-by-default TV renamer using TVDB v4.

## Quick start
- Set `TVDB_APIKEY` (and `TVDB_PIN` if needed)
- Create `~/.tvrn/config.toml` if you want non-defaults
- From a series/season folder, run: `tvrn`
- Review preview, type `Y` to apply

## Defaults
- Order: aired
- Lang: en
- Multi-ep: range (S01E01-E02)
- Strict confirmation: only `Y` proceeds

## Roadmap
- Implement TVDB client calls and ETag-aware cache
- Parser coverage for common scene naming
- Atomic-ish batch commit with EXDEV handling