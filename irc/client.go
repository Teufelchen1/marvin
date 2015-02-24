// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package irc

import (
	"fmt"
	"net"
	"strings"
)

type Hook func(*Client, Message) error

type Client struct {
	conn     *net.Conn
	hooks    map[string][]Hook
	Channels []string
}

func NewClient(conn *net.Conn) *Client {
	c := &Client{
		conn:  conn,
		hooks: make(map[string][]Hook),
	}

	c.CmdHook("ping", c.pingCmd)
	c.CmdHook("join", c.joinCmd)
	c.CmdHook("kick", c.kickCmd)
	c.CmdHook("quit", c.quitCmd)

	return c
}

func (c *Client) pingCmd(client *Client, msg Message) error {
	return c.Write("PONG %s", msg.Data)
}

func (c *Client) joinCmd(client *Client, msg Message) error {
	c.Channels = append(c.Channels, msg.Data)
	return nil
}

func (c *Client) kickCmd(client *Client, msg Message) error {
	var newChannels []string
	channel := strings.Fields(msg.Receiver)[0]

	for _, ch := range c.Channels {
		if ch != channel {
			newChannels = append(newChannels, ch)
		}
	}

	c.Channels = newChannels
	return nil
}

func (c *Client) quitCmd(client *Client, msg Message) error {
	client.Channels = []string{}
	return nil
}

func (c *Client) Write(format string, argv ...interface{}) error {
	_, err := fmt.Fprintf(*c.conn, "%s\r\n", fmt.Sprintf(format, argv...))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Handle(data string, ch chan error) {
	msg := parseMessage(data)
	hooks, ok := c.hooks[msg.Command]
	if ok {
		for _, hook := range hooks {
			go func(h Hook) {
				if err := h(c, msg); err != nil {
					ch <- err
				}
			}(hook)
		}
	}
}

func (c *Client) CmdHook(cmd string, hook Hook) {
	c.hooks[cmd] = append(c.hooks[cmd], hook)
}
