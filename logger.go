package bitcoinaverage

type logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
}

func (c *Client) log(a ...interface{}) {
	if c.logger != nil {
		c.logger.Print(a...)
	}
}

func (c *Client) logf(f string, a ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(f, a...)
	}
}

func (c *Client) debug(a ...interface{}) {
	if c.logger != nil {
		c.logger.Debug(a...)
	}
}

func (c *Client) debugf(f string, a ...interface{}) {
	if c.logger != nil {
		c.logger.Debugf(f, a...)
	}
}
