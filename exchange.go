package bitcoinaverage

// Exchange is the full data for a single exchange
type Exchange struct {
	Name        string                   `json:"name"`
	DisplayName string                   `json:"display_name"`
	URL         string                   `json:"url"`
	Timestamp   int64                    `json:"timestamp"`
	DataSource  string                   `json:"data_source"`
	Symbols     map[Pair]*ExchangeSymbol `json:"symbols"`
	Success     bool                     `json:"success"`
}

// ExchangeSymbol is the data for an exhcnage's symbol
type ExchangeSymbol struct {
	Last   float64 `json:"last"`
	Volume float64 `json:"volume"`
	Ask    float64 `json:"ask"`
	Bid    float64 `json:"bid"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Open   float64 `json:"open"`
	Vwap   float64 `json:"vwap"`
}
