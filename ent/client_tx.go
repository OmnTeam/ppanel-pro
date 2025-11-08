package ent

import (
	"context"
	"fmt"
	"log"
)

func (c *Client) TX(ctx context.Context, fn func(tx *Tx) error) (err error) {
	tx, err := c.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			log.Printf("panic in transaction: %v\n", v)
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, err2)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
