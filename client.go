package bitcoinaverage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _ = bytes.MinRead

const (
	apiURL = "apiv2.bitcoinaverage.com"
)

// Client is a client to connect to the bitcoinaverage API
type Client struct {
	publicKey, privateKey string
	http                  http.Client
	logger                logger
}

// NewClient returns a new Client instance with the keys set
func NewClient(publicKey, privateKey string, l logger) *Client {
	return &Client{
		publicKey:  publicKey,
		privateKey: privateKey,
		http:       http.Client{},
		logger:     l,
	}
}

func (c *Client) getSignature() string {
	// get timestamp
	t := time.Now()
	ts := int(t.Unix())
	timestamp := strconv.Itoa(ts)

	payload := timestamp + "." + c.publicKey

	// prepare the hmac with sha256
	hash := hmac.New(sha256.New, []byte(c.privateKey))
	hash.Write([]byte(payload))
	// hex representation
	hexValue := hex.EncodeToString(hash.Sum(nil))

	return payload + "." + hexValue
}

func makeReqURL(scheme, path string, params url.Values) string {
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	reqURL := fmt.Sprintf("%s://%s%s", scheme, apiURL, path)
	if params != nil {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}
	return reqURL
}

func (c *Client) doReq(path string, params url.Values) (*http.Response, error) {
	reqURL := makeReqURL("https", path, params)
	r, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "making request for %s", path)
	}
	r.Header.Set("X-signature", c.getSignature())
	res, err := c.http.Do(r)
	if err != nil {
		return nil, errors.Wrapf(err, "doing request for %s", path)
	}
	if res.StatusCode >= 400 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "reading response for %s", path)
		}
		return nil, errors.Errorf("[%d] %s from bitcoinaverage API: %s", res.StatusCode, res.Status, bodyBytes)
	}
	return res, nil
}

func (c *Client) getWebsocketTicket() (*WebsocketTicket, error) {
	res, err := c.doReq("websocket/get_ticket", nil)
	if err != nil {
		return nil, err
	}
	// bodyBytes, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// c.debug(string(bodyBytes))
	// res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	dec := json.NewDecoder(res.Body)
	ticket := &WebsocketTicket{}
	err = dec.Decode(ticket)
	if err != nil {
		return nil, errors.Wrap(err, "decoding JSON")
	}
	return ticket, nil
}

func (c *Client) monitorTickerStream(conn *websocket.Conn, dataChan chan *Ticker, errChan chan error, stopChan chan bool) {
	c.debug("monitoring websocket")
	defer conn.Close()
	for {
		select {
		case <-stopChan:
			close(dataChan)
			close(errChan)
			return
		default:
			c.debug("reading from websocket")
			resp := &WebsocketTicker{}
			err := conn.ReadJSON(resp)
			if err != nil {
				errChan <- errors.Wrap(err, "reading JSON")
			} else {
				c.debugf("got ticker data from read: %#v", resp.Data)
				dataChan <- resp.Data
			}
		}
	}
}

func (c *Client) monitorExchangeStream(conn *websocket.Conn, dataChan chan *Exchange, errChan chan error, stopChan chan bool) {
	c.debug("monitoring websocket")
	defer conn.Close()
	for {
		select {
		case <-stopChan:
			close(dataChan)
			close(errChan)
			return
		default:
			c.debug("reading from websocket")
			resp := &WebsocketExchange{}
			err := conn.ReadJSON(resp)
			if err != nil {
				errChan <- errors.Wrap(err, "reading JSON")
			} else {
				c.debugf("got ticker data from read: %#v", resp.Data)
				dataChan <- resp.Data
			}
		}
	}
}

// Exchanges returns the global ecxhange data for the supplied cryptos and fiats
func (c *Client) Exchanges(cryptos, fiats []string) ([]*Exchange, error) {
	params := url.Values{}

	if cryptos != nil && len(cryptos) > 0 {
		params.Set("crypto", strings.Join(cryptos, ","))
	}
	if fiats != nil && len(fiats) > 0 {
		params.Set("fiat", strings.Join(fiats, ","))
	}

	res, err := c.doReq("/exchanges/ticker/all", params)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(res.Body)
	data := []*Exchange{}
	if err := dec.Decode(&data); err != nil {
		return nil, errors.Wrap(err, "decoding JSON")
	}
	return data, nil
}

// ExchangeStream will stream one or more exchanges
func (c *Client) ExchangeStream(exchanges ...string) (chan *Exchange, chan error, chan bool, error) {
	dataChan := make(chan *Exchange, 2)
	errChan := make(chan error)
	stopChan := make(chan bool)

	conn, err := c.getSocketConnection("websocket/multiple/exchanges")
	if err != nil {
		return nil, nil, nil, err
	}

	for _, e := range exchanges {
		c.debugf("got socket connection %s -> %s", conn.LocalAddr(), conn.RemoteAddr())
		err = conn.WriteJSON(&WebsocketCommand{
			Event: "message",
			Data: &WebsocketOperation{
				Operation: "subscribe",
				Options: &WebsocketOperationOptions{
					Exchange: e,
				},
			},
		})
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "writing subscribe messages")
		}
		resp := &WebsocketCommandResponse{}
		err = conn.ReadJSON(resp)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "reading JSON from websocket")
		}

		if resp.Data != "OK" {
			return nil, nil, nil, fmt.Errorf("Non OK command response: %s", resp.Data)
		}

	}

	go c.monitorExchangeStream(conn, dataChan, errChan, stopChan)

	return dataChan, errChan, stopChan, nil
}

func (c *Client) getSocketConnection(urlStr string) (*websocket.Conn, error) {

	ticket, err := c.getWebsocketTicket()
	if err != nil {
		return nil, err
	}
	c.debug("got socket ticket")

	params := url.Values{}
	params.Add("ticket", ticket.Ticket)
	params.Add("public_key", c.publicKey)
	socketURL := makeReqURL("wss", urlStr, params)
	c.debug("connecting to socket", socketURL)
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "dialing socket to %s", socketURL)
	}
	return conn, nil

}

// Tickers returns the global ticker data for the supplied cryptos and fiats
func (c *Client) Tickers(cryptos, fiats []string) (MultiTicker, error) {
	params := url.Values{}

	if cryptos != nil && len(cryptos) > 0 {
		params.Set("crypto", strings.Join(cryptos, ","))
	}
	if fiats != nil && len(fiats) > 0 {
		params.Set("fiat", strings.Join(fiats, ","))
	}

	res, err := c.doReq("/indices/global/ticker/all", params)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(res.Body)
	data := MultiTicker{}
	if err := dec.Decode(&data); err != nil {
		return nil, errors.Wrap(err, "decoding JSON")
	}
	return data, nil
}

// TickerStream will send *Ticker data on the supplied dataChan
func (c *Client) TickerStream(tickers ...string) (chan *Ticker, chan error, chan bool, error) {
	dataChan := make(chan *Ticker, 2)
	errChan := make(chan error)
	stopChan := make(chan bool)

	conn, err := c.getSocketConnection("websocket/multiple/ticker")
	if err != nil {
		return nil, nil, nil, err
	}

	c.debugf("got socket connection %s -> %s", conn.LocalAddr(), conn.RemoteAddr())
	for _, ticker := range tickers {
		err = conn.WriteJSON(&WebsocketCommand{
			Event: "message",
			Data: &WebsocketOperation{
				Operation: "subscribe",
				Options: &WebsocketOperationOptions{
					Currency:  ticker,
					SymbolSet: "global",
				},
			},
		})
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "writing subscribe message to socket")
		}

		resp := &WebsocketCommandResponse{}
		err = conn.ReadJSON(resp)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "reading JSON from websocket")
		}

		if resp.Data != "OK" {
			return nil, nil, nil, fmt.Errorf("Non OK command response: %s", resp.Data)
		}

	}

	go c.monitorTickerStream(conn, dataChan, errChan, stopChan)

	return dataChan, errChan, stopChan, nil
}
