package cart

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	sqlCreateCart   = `INSERT INTO carts (user_id, created_at, updated_at) VALUES ($1, $2, $3) RETURNING cart_id`
	sqlDeleteCart   = `DELETE FROM carts WHERE cart_id = $1`
	sqlCartByID     = `SELECT user_id, created_at, updated_at FROM carts WHERE cart_id = $1`
	sqlUpdateCartTS = `UPDATE carts SET updated_at = $2 WHERE cart_id = $1`

	sqlLinesByCartID   = `SELECT product_id, quantity FROM line_items WHERE cart_id = $1`
	sqlProductQuantity = `SELECT quantity FROM line_items WHERE cart_id = $1 AND product_id = $2`

	sqlCreateLineItem  = `INSERT INTO line_items (cart_id, product_id, quantity, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	sqlUpdateLineItem  = `UPDATE line_items SET quantity = $3, updated_at = $4 WHERE cart_id = $1 AND product_id = $2`
	sqlDeleteLineItem  = `DELETE FROM line_items WHERE cart_id = $1 AND product_id = $2`
	sqlDeleteLineItems = `DELETE FROM line_items WHERE cart_id = $1`
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

var readOnly = &sql.TxOptions{ReadOnly: true}

func lineItem(ctx context.Context, tx *sql.Tx, cartID, productID int64) (LineItem, error) {
	li := LineItem{
		ProductID: productID,
	}

	row := tx.QueryRowContext(ctx, sqlProductQuantity, cartID, productID)
	if err := row.Scan(&li.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return LineItem{}, errNotFound
		}
		return LineItem{}, err
	}

	return li, nil
}

func createLineItem(ctx context.Context, tx *sql.Tx, cartID int64, li LineItem) error {
	_, err := tx.ExecContext(ctx, sqlCreateLineItem, cartID, li.ProductID, li.Quantity, li.CreatedAt, li.UpdatedAt)
	return err
}

func updateLineItem(ctx context.Context, tx *sql.Tx, cartID int64, li LineItem) error {
	_, err := tx.ExecContext(ctx, sqlUpdateLineItem, cartID, li.ProductID, li.Quantity, li.UpdatedAt)
	return err
}

func (s *Storage) AddProduct(ctx context.Context, cartID, productID int64, quantity uint32) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %s", err)
	}

	exists := true

	li, err := lineItem(ctx, tx, cartID, productID)
	if err != nil {
		if err == errNotFound {
			exists = false
		} else {
			if err := tx.Rollback(); err != nil {
				log.Printf("Rollback failed: %s", err)
			}
			return err
		}
	}

	now := time.Now()

	if exists {
		li.Quantity += quantity
		li.UpdatedAt = now
		err = updateLineItem(ctx, tx, cartID, li)
	} else {
		li = LineItem{
			ProductID: productID,
			Quantity:  quantity,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = createLineItem(ctx, tx, cartID, li)
	}

	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback failed: %s", err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	return nil
}

func (s *Storage) DeleteProduct(ctx context.Context, cartID, productID int64) error {
	if _, err := s.db.ExecContext(ctx, sqlDeleteLineItem, cartID, productID); err != nil {
		if err == sql.ErrNoRows {
			return errNotFound
		}

		return err
	}

	return nil
}

func (s *Storage) CartByID(ctx context.Context, id int64) (Cart, error) {
	tx, err := s.db.BeginTx(ctx, readOnly)
	if err != nil {
		return Cart{}, fmt.Errorf("failed to start transaction: %s", err)
	}

	cart := Cart{ID: id}

	row := tx.QueryRowContext(ctx, sqlCartByID, id)
	if err := row.Scan(&cart.UserID, &cart.CreatedAt, &cart.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Cart{}, errNotFound
		}

		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback failed: %s", err)
		}
		return Cart{}, err
	}

	rows, err := tx.QueryContext(ctx, sqlLinesByCartID, id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback failed: %s", err)
		}

		if err == sql.ErrNoRows {
			return Cart{}, errNotFound
		}

		return Cart{}, err
	}
	defer rows.Close()

	for rows.Next() {
		li := LineItem{}
		if err := rows.Scan(&li.ProductID, &li.Quantity); err != nil {
			return Cart{}, fmt.Errorf("failed to scan row into LineItem sruct: %s", err)
		}
		cart.Items = append(cart.Items, li)
	}

	if err := rows.Err(); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback failed: %s", err)
		}

		return Cart{}, fmt.Errorf("failed to iterate over DB rows: %s", err)
	}

	if err := tx.Commit(); err != nil {
		return Cart{}, fmt.Errorf("failed to commit transaction: %s", err)
	}

	return cart, nil
}

func (s *Storage) CreateCart(ctx context.Context, cart Cart) (Cart, error) {
	err := s.db.QueryRowContext(ctx, sqlCreateCart, cart.UserID, cart.CreatedAt, cart.UpdatedAt).Scan(&cart.ID)
	if err != nil {
		return Cart{}, err
	}

	return cart, nil
}

func deleteLineItems(ctx context.Context, tx *sql.Tx, cartID int64) error {
	if _, err := tx.ExecContext(ctx, sqlDeleteLineItems, cartID); err != nil {
		if err == sql.ErrNoRows {
			return errNotFound
		}

		return err
	}

	return nil
}

func (s *Storage) DeleteCart(ctx context.Context, cartID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %s", err)
	}

	if err := deleteLineItems(ctx, tx, cartID); err != nil {
		if err != errNotFound {
			if err := tx.Rollback(); err != nil {
				log.Printf("Rollback failed: %s", err)
			}
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, sqlDeleteCart, cartID); err != nil {
		if err != sql.ErrNoRows {
			if err := tx.Rollback(); err != nil {
				log.Printf("Rollback failed: %s", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	return nil
}

func (s *Storage) DeleteLineItems(ctx context.Context, cartID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %s", err)
	}

	if err := deleteLineItems(ctx, tx, cartID); err != nil {
		if err != errNotFound {
			if err := tx.Rollback(); err != nil {
				log.Printf("Rollback failed: %s", err)
			}
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, sqlUpdateCartTS, cartID, time.Now()); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback failed: %s", err)
		}

		if err == sql.ErrNoRows {
			return errNotFound
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	return nil
}
