package smtp

import (
	"net"
	"net/textproto"
)

type Client struct {
	conn       net.Conn
	text       *textproto.Conn
	serverName string
	lmtp       bool
	exec       map[string]string
}
