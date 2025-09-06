package planner

type Options struct {
  Scheme   string // XxYY, SXXEYY, etc
  Pad      int
  Order    string // aired|dvd|absolute
  Lang     string
  Specials string // ignore|inline|folder
  MultiEP  string // range|join
}

type Item struct {
  From     string
  To       string
  Reason   string // e.g. rename, collision-skip
}

type Plan struct {
  Items []Item
}

type Stats struct {
  Total int
  Collisions int
  Skipped int
}
