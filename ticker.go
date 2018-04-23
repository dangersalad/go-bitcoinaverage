package bitcoinaverage

import (
	"encoding/json"
)

// MultiTicker is Ticker data for multiple symbols
type MultiTicker map[Pair]*Ticker

// MultiSet is Ticker data for multiple symbols in multiple symbol sets
type MultiSet map[string]*MultiTicker

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
	RawDay   json.Number `json:"day"`
	RawWeek  json.Number `json:"week"`
	RawMonth json.Number `json:"month"`
}

// Day returns the float64 value for the potential number passed in, 0 if it was not a number
func (d *DWM) Day() float64 {
	f, err := d.RawDay.Float64()
	if err != nil {
		return 0
	}
	return f
}

// GetDay returns the float64 value for the potential number passed in, error if it was not a number
func (d *DWM) GetDay() (float64, error) {
	return d.RawDay.Float64()
}

// Week returns the float64 value for the potential number passed in, 0 if it was not a number
func (d *DWM) Week() float64 {
	f, err := d.RawWeek.Float64()
	if err != nil {
		return 0
	}
	return f
}

// GetWeek returns the float64 value for the potential number passed in, error if it was not a number
func (d *DWM) GetWeek() (float64, error) {
	return d.RawWeek.Float64()
}

// Month returns the float64 value for the potential number passed in, 0 if it was not a number
func (d *DWM) Month() float64 {
	f, err := d.RawMonth.Float64()
	if err != nil {
		return 0
	}
	return f
}

// GetMonth returns the float64 value for the potential number passed in, error if it was not a number
func (d *DWM) GetMonth() (float64, error) {
	return d.RawMonth.Float64()
}

// Changes holds the percent and price changes ising DWM stucts
type Changes struct {
	Percent *DWM `json:"percent"`
	Price   *DWM `json:"price"`
}
