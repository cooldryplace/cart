package cart

import (
	"context"
	"errors"
	"log"
	"time"
)

var (
	errNotImplemented = errors.New("not implemented")
	errNotFound       = errors.New("not found")
)

type storage interface {
	AddProduct(ctx context.Context, cartID, productID int64, quantity uint32) error
	DeleteProduct(ctx context.Context, cartID, productID int64) error
	CartByID(ctx context.Context, id int64) (Cart, error)
	CreateCart(ctx context.Context, cart Cart) (Cart, error)
	DeleteCart(ctx context.Context, cartID int64) error
	DeleteLineItems(ctx context.Context, cartID int64) error
}

// Carts contains all business logic realated to this microservice.
type Carts struct {
	storage storage
}

// New builds and returns new instance of Carts that is ready for use.
func New(s storage) *Carts {
	return &Carts{storage: s}
}

// LineItem represents single SKU and quantity.
type LineItem struct {
	ProductID int64
	Quantity  uint32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Cart holds LineItems for User.
type Cart struct {
	ID        int64
	UserID    int64
	Items     []LineItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AddProduct to a Cart.
func (c *Carts) AddProduct(ctx context.Context, cartID, productID int64, quantity uint32) error {
	if err := c.storage.AddProduct(ctx, cartID, productID, quantity); err != nil {
		log.Printf("Failed to add a Product: %d to the Cart: %d, error: %s", productID, cartID, err)
		return err
	}

	return nil
}

// DeleteProduct in a Cart.
func (c *Carts) DeleteProduct(ctx context.Context, cartID, productID int64) error {
	if err := c.storage.DeleteProduct(ctx, cartID, productID); err != nil {
		log.Printf("Failed to delete the Product: %d from the Cart: %d, error: %s", productID, cartID, err)
		return err
	}

	return nil
}

// Cart returns Cart with provided ID.
func (c *Carts) Cart(ctx context.Context, id int64) (Cart, error) {
	cart, err := c.storage.CartByID(ctx, id)
	if err != nil {
		if err != errNotFound {
			log.Printf("Failed to get the Cart with ID: %d, error: %s", id, err)
		}
		return Cart{}, err
	}

	return cart, nil
}

// Create Cart for a User.
func (c *Carts) Create(ctx context.Context, userID int64) (Cart, error) {
	now := time.Now()

	cart := Cart{
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	cart, err := c.storage.CreateCart(ctx, cart)
	if err != nil {
		log.Printf("Failed to create a Cart for UserID: %d, error: %s", userID, err)
		return Cart{}, err
	}

	return cart, nil
}

func (c *Carts) Delete(ctx context.Context, cartID int64) error {
	if err := c.storage.DeleteCart(ctx, cartID); err != nil {
		log.Printf("Failed to delete the Cart with ID: %d, error: %s", cartID, err)
		return err
	}

	return nil
}

func (c *Carts) Empty(ctx context.Context, cartID int64) error {
	if err := c.storage.DeleteLineItems(ctx, cartID); err != nil {
		log.Printf("Failed to empty the Cart with ID: %d, error: %s", cartID, err)
		return err
	}

	return nil
}
