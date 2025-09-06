# TVRN — TV Renamer

Fast, safe-by-default TV file renamer that uses TheTVDB v4 for metadata lookups

* Aired order by default with order overrides
* Multi-episode aware with clean range formatting (`1x01-02 - Title1 + Title2`)
* Shows a dry-run proposal first and requires explicit confirmation
* Skips files already named correctly, and refuses to proceed when the season doesn’t exist in the selected order
* Caches API responses locally for speed

> You provide your own TVDB API key. This project does not distribute media

## Install

### Go users

```
go install github.com/GizzmoShifu/tvrn/cmd/tvrn@latest
```

Ensure `$(go env GOBIN)` is on your `PATH`, then run `tvrn -h`

### From source (repo checkout)

```
go mod tidy
go build -trimpath -ldflags="-s -w" -o ~/.local/bin/tvrn ./cmd/tvrn
```

Add `~/.local/bin` to `PATH` if needed

## Configure

Create `~/.tvrn/config.toml` (recommended) or set `TVDB_APIKEY` in your shell

```bash
export TVDB_APIKEY=YOUR_TVDB_API_KEY
```

OR

```toml
# ~/.tvrn/config.toml

[auth]
apikey = "YOUR_TVDB_API_KEY"
pin    = ""                # optional

[defaults]
order  = "aired"           # aired | dvd | absolute | alternate | regional
lang   = "en"              # title language
confirmation_strict = true # only capital Y proceeds

[rename]
scheme   = "XxYY"          # SXXEYY | sXXeYY | XxYY | XYY | YY
pad      = 2               # digits to pad episode number
multi_ep = "range"         # range | join
```

Local cache lives in `~/.tvrn/cache`

## Usage

```
tvrn [options] [path]
```

If you run `tvrn` inside a season folder (e.g. `.../Show Name/Season 1`) it will

* Detect the series and season from the path
* Query TVDB for that season using your configured order
* Propose new filenames, sorted by episode
* Ask for confirmation — anything but `Y/y/yes` cancels

### Common options

* `--scheme` set episode format
  `SXXEYY` | `sXXeYY` | `XxYY` | `XYY` | `YY`
* `--pad` pad episode number to N digits
  default 2
* `--order` episode order used for metadata lookup
  `aired` | `dvd` | `absolute` | `alternate` | `regional`
* `--lang` episode title language
  default `en`
* `--multi` multi-episode naming
  `range` uses `1x01-02`, `join` uses `1x01x02`
* `--detailed` show `before -> after` in the proposal
* `--debug` verbose matching and API traces
* `--series` run from a series root and process all “Season \*” subfolders
* `--no-cache` ignore local API cache for this run
* `--yes` auto-confirm for non-interactive runs
* `--about` show credits and licensing notices and exit
* `--version` show version metadata and exit

### Examples

* In a season folder
  `tvrn`

* From series root for all season subfolders
  `tvrn --series`

* DVD order with explicit formatting and detailed preview
  `tvrn --order=dvd --scheme=SXXEYY --pad=2 --detailed`

## Behaviour

* **Safe by default**
  Always previews and asks for `Y` unless `--yes` is set

* **No-op skips**
  If the destination name already equals the source, it’s skipped and not shown in the plan

* **Season checks**
  If the selected season has no episodes in the chosen order, the run fails early and lists the seasons TVDB does have for that series

* **Unknown episode numbers**
  Files that refer to episode numbers missing in TVDB for that season are skipped. If all files are unknown, the run exits with a clear error

* **Multi-episode formatting**
  In `range` mode, the second number does not repeat the prefix
  `1x01-02 - Title1 + Title2.ext`
  `S01E01-E02 - Title1 + Title2.ext`

* **Sorting**
  The proposal is shown in S/E order so it’s easy to eyeball

## Caching

* Location
  `~/.tvrn/cache`

* Episodes cache key
  `episodes:{seriesID}:{order}:{season}:{lang}.json`

* TTL
  Defaults to roughly a day per page. Use `--no-cache` to bypass for a run

## Troubleshooting

* **Series found, but “no episodes for season N”**
  Your folder or files are using a season number different to TVDB’s for that order. Rename the folder/files to TVDB’s season numbering and re-run

* **Titles missing for a file**
  We only rename when we can match the episode number(s) for the season. Unknown episodes are skipped and reported

* **Specials**
  Season `0` is supported by TVDB. If you file specials separately, run `tvrn` inside the `Specials` folder

* **Rate limits**
  The client backs off on HTTP 429. If you script large runs, consider short sleeps between runs

## Attribution

This product uses the [TheTVDB.com](https://thetvdb.com) v4 API for metadata
It is not endorsed or certified by TheTVDB

If you rely on this data, please consider contributing updates directly on TheTVDB

## Roadmap

* Collision policy: skip | suffix | overwrite, configurable
* Windows-safe sanitising for trailing dots and spaces
* Per-series overrides in config for known quirks
* Unit tests for the parser and formatter

Contributions and issue reports are welcome

## Ethical use

* You must provide your own TVDB API key and comply with TheTVDB terms
* This project does not include or distribute media
* Intended for metadata hygiene and personal library organisation

## Licence

MIT — see `LICENSE` in this repository
