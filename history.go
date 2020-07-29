package bitcoinaverage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// HistoryData is the full data for a single exchange
type HistoryData struct {
	Average float64 `json:"average"`
	Time    string  `json:"time"`
}

type HistoryResolution string

const (
	HistoryResolutionMinute HistoryResolution = "minute"
	HistoryResolutionHour   HistoryResolution = "hour"
	HistoryResolutionDay    HistoryResolution = "day"
)

// PriceAtTimestamp returns the global ticker data for the supplied cryptos and fiats
func (c *Client) PriceAtTimestamp(symbol Pair, at time.Time, resolution HistoryResolution) (*HistoryData, error) {
	params := url.Values{}

	params.Set("at", strconv.FormatInt(at.Unix(), 10))
	params.Set("resolution", string(resolution))

	path := fmt.Sprintf("/indices/global/history/%s", symbol)

	res, err := c.doReq(path, params)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(res.Body)
	data := HistoryData{}
	if err := dec.Decode(&data); err != nil {
		return nil, errors.Wrap(err, "decoding JSON")
	}

	return &data, nil
}
