package ldaps

import (
	"context"
)

// Ping checks the LDAP dependency is reachable and bind works.
func (c *Client) Ping(ctx context.Context) error {
	conn, err := c.dialAndBind(ctx)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
