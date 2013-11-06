// Copyright 2013 The Pythia Authors.
// This file is part of Pythia.
//
// Pythia is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// Pythia is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

package pythia

import (
	"encoding/json"
	"io"
	"log"
	"net"
)

// BUG(vianney): When a connection is closed from the remote side, the writer
// goroutine of pythia.Conn will remain running.

// Conn is a wrapper over net.Conn, reading and writing Messages.
type Conn struct {
	// The underlying connection.
	conn net.Conn

	// Incoming messages channel.
	input chan Message

	// Flag to ignore errors after closing the connection.
	closed bool
}

// WrapConn wraps a stream network connection into a Message-oriented
// connection. The raw conn connection shall not be used by the user anymore.
func WrapConn(conn net.Conn) *Conn {
	c := new(Conn)
	c.conn = conn
	c.input = make(chan Message)
	go c.reader()
	return c
}

// The reader goroutine parses the Messages and put them in the input channel.
func (c *Conn) reader() {
	defer close(c.input)
	dec := json.NewDecoder(c.conn)
	for {
		var msg Message
		if err := dec.Decode(&msg); err == io.EOF || c.closed {
			break
		} else if err != nil {
			log.Print(err)
		}
		c.input <- msg
	}
}

// Dial connects to the address addr and returns a Message-oriented connection.
func Dial(addr net.Addr) (*Conn, error) {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return WrapConn(conn), nil
}

// Send sends a message through the connection.
func (c *Conn) Send(msg Message) error {
	enc := json.NewEncoder(c.conn)
	return enc.Encode(msg)
}

// Receive returns the channel from which incoming messages can be retrieved.
func (c *Conn) Receive() <-chan Message {
	return c.input
}

// Close closes the connection. The receive channel will also be closed.
func (c *Conn) Close() error {
	c.closed = true
	return c.conn.Close()
}

// vim:set sw=4 ts=4 noet: