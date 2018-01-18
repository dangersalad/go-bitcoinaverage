package bitcoinaverage

// WebsocketCommandResponse is a message from the bitcoinaverae
// websocket in response to a WebsocketCommand
type WebsocketCommandResponse struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// WebsocketTicker holds the ticker data wrapped in the websocket JSON
type WebsocketTicker struct {
	Event string  `json:"event"`
	Data  *Ticker `json:"data"`
}

// WebsocketTicket is the auth token used to connect the socket
type WebsocketTicket struct {
	Ticket string `json:"ticket"`
}

// WebsocketCommand subscribes or unsubscribes from the ticker socket
type WebsocketCommand struct {
	Event string              `json:"event"`
	Data  *WebsocketOperation `json:"data"`
}

// WebsocketOperation is the actual command being sent in a WebsocketCommand
type WebsocketOperation struct {
	Operation string                     `json:"operation"`
	Options   *WebsocketOperationOptions `json:"options"`
}

// WebsocketOperationOptions are options for an operation
type WebsocketOperationOptions struct {
	Currency  string `json:"currency"`
	SymbolSet string `json:"symbol_set"`
}
