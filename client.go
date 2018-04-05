package bitcoinaverage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var _ = log.Print
var _ = bytes.MinRead

const (
	apiURL = "apiv2.bitcoinaverage.com"
)

// Client is a client to connect to the bitcoinaverage API
type Client struct {
	publicKey, privateKey string
	http                  http.Client
}

// NewClient returns a new Client instance with the keys set
func NewClient(publicKey, privateKey string) *Client {
	return &Client{
		publicKey:  publicKey,
		privateKey: privateKey,
		http:       http.Client{},
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
		return nil, err
	}
	r.Header.Set("X-signature", c.getSignature())
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("[%d] %s from bitcoinaverage API: %s", res.StatusCode, res.Status, bodyBytes)
	}
	return res, nil
}

func (c *Client) getWebsocketTicket() (*WebsocketTicket, error) {
	res, err := c.doReq("websocket/get_ticket", nil)
	// bodyBytes, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// log.Println(string(bodyBytes))
	// res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	dec := json.NewDecoder(res.Body)
	ticket := &WebsocketTicket{}
	err = dec.Decode(ticket)
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func monitorTickerStream(conn *websocket.Conn, dataChan chan *Ticker, errChan chan error, stopChan chan bool) {
	// log.Println("monitoring websocket")
	defer conn.Close()
	for {
		select {
		case <-stopChan:
			close(dataChan)
			close(errChan)
			return
		default:
			// log.Println("reading from websocket")
			resp := &WebsocketTicker{}
			err := conn.ReadJSON(resp)
			if err != nil {
				// log.Println("got error from read", err)
				errChan <- err
			} else {
				// log.Printf("got ticker data from read: %#v", resp.Data)
				dataChan <- resp.Data
			}
		}
	}
}

func monitorExchangeStream(conn *websocket.Conn, dataChan chan *Exchange, errChan chan error, stopChan chan bool) {
	// log.Println("monitoring websocket")
	defer conn.Close()
	for {
		select {
		case <-stopChan:
			close(dataChan)
			close(errChan)
			return
		default:
			// log.Println("reading from websocket")
			resp := &WebsocketExchange{}
			err := conn.ReadJSON(resp)
			if err != nil {
				// log.Println("got error from read", err)
				errChan <- err
			} else {
				// log.Printf("got ticker data from read: %#v", resp.Data)
				dataChan <- resp.Data
			}
		}
	}
}

// ExchangeStream will stream one or more exchanges
func (c *Client) ExchangeStream(exchanges ...string) (chan *Exchange, chan error, chan bool, error) {
	dataChan := make(chan *Exchange, 2)
	errChan := make(chan error)
	stopChan := make(chan bool)

	conn, err := c.getSocketConnection("websocket/exchanges")
	if err != nil {
		return nil, nil, nil, err
	}

	for _, e := range exchanges {
		// log.Printf("got socket connection %s -> %s", conn.LocalAddr(), conn.RemoteAddr())
		err = conn.WriteJSON(&WebsocketCommand{
			Event: "message",
			Data: &WebsocketOperation{
				Operation: "subscribe",
				Options: &WebsocketOperationOptions{
					Exchange: e,
				},
			},
		})
	}

	if err != nil {
		return nil, nil, nil, err
	}

	resp := &WebsocketCommandResponse{}
	err = conn.ReadJSON(resp)
	if err != nil {
		return nil, nil, nil, err
	}

	if resp.Data != "OK" {
		return nil, nil, nil, fmt.Errorf("Non OK command response: %s", resp.Data)
	}

	go monitorExchangeStream(conn, dataChan, errChan, stopChan)

	return dataChan, errChan, stopChan, nil
}

func (c *Client) getSocketConnection(urlStr string) (*websocket.Conn, error) {

	ticket, err := c.getWebsocketTicket()
	if err != nil {
		return nil, err
	}
	// log.Println("got socket ticket")

	params := url.Values{}
	params.Add("ticket", ticket.Ticket)
	params.Add("public_key", c.publicKey)
	socketURL := makeReqURL("wss", urlStr, params)
	// log.Println("connecting to socket", socketURL)
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil

}

// TickerStream will send *Ticker data on the supplied dataChan
func (c *Client) TickerStream(ticker string) (chan *Ticker, chan error, chan bool, error) {
	dataChan := make(chan *Ticker, 2)
	errChan := make(chan error)
	stopChan := make(chan bool)

	conn, err := c.getSocketConnection("websocket/ticker")
	if err != nil {
		return nil, nil, nil, err
	}

	// log.Printf("got socket connection %s -> %s", conn.LocalAddr(), conn.RemoteAddr())
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
		return nil, nil, nil, err
	}

	resp := &WebsocketCommandResponse{}
	err = conn.ReadJSON(resp)
	if err != nil {
		return nil, nil, nil, err
	}

	if resp.Data != "OK" {
		return nil, nil, nil, fmt.Errorf("Non OK command response: %s", resp.Data)
	}

	go monitorTickerStream(conn, dataChan, errChan, stopChan)

	return dataChan, errChan, stopChan, nil
}
