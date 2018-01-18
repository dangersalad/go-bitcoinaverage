package bitcoinaverage

// Ticker is the full data for a single ticker
type Ticker struct {
	Ask              float64  `json:"ask"`
	Bid              float64  `json:"bid"`
	Last             float64  `json:"last"`
	High             float64  `json:"high"`
	Low              float64  `json:"low"`
	Open             *DWM     `json:"open"`
	Averages         *DWM     `json:"averages"`
	Changes          *Changes `json:"changes"`
	Volume           float64  `json:"volume"`
	VolumePercent    float64  `json:"volume_percent"`
	Timestamp        int64    `json:"timestamp"`
	DislpayTimestamp string   `json:"display_timestamp"`
	Success          bool     `json:"success"`
	Time             string   `json:"time"`
}

// DWM holds data for a set of float values dor "day", "week", and "month"
type DWM struct {
	Day   float64 `json:"day"`
	Week  float64 `json:"week"`
	Month float64 `json:"month"`
}

// Changes holds the percent and price changes ising DWM stucts
type Changes struct {
	Percent *DWM `json:"percent"`
	Price   *DWM `json:"price"`
}
