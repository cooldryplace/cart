package cart

import (
	"context"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	var (
		generatedID    int64 = 99
		expectedUserID int64 = 13
		start                = time.Now()
	)

	storage := &StorageMock{
		CreateCartFunc: func(ctx context.Context, cart Cart) (Cart, error) {
			end := time.Now()

			if cart.UserID != expectedUserID {
				t.Errorf("Got userID: %d, expected: %d", cart.UserID, expectedUserID)
			}

			if cart.CreatedAt != cart.UpdatedAt {
				t.Error("CreatedAt not equal to UpdatedAt")
			}

			if cart.CreatedAt.Before(start) {
				t.Error("Cart timestamps are to low")
			}

			if cart.CreatedAt.After(end) {
				t.Error("Cart timestamps are to high")
			}

			cart.ID = generatedID

			return cart, nil
		},
	}

	carts := New(storage)

	actual, err := carts.Create(context.Background(), expectedUserID)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if actual.ID != generatedID {
		t.Errorf("Got ID: %d, expected: %d", actual.ID, generatedID)
	}
}

// StorageMock allows you dinamically set Storage behavior.
type StorageMock struct {
	AddProductFunc      func(ctx context.Context, cartID, productID int64, quantity uint32) error
	DeleteProductFunc   func(ctx context.Context, cartID, productID int64) error
	CartByIDFunc        func(ctx context.Context, id int64) (Cart, error)
	CreateCartFunc      func(ctx context.Context, cart Cart) (Cart, error)
	DeleteCartFunc      func(ctx context.Context, cartID int64) error
	DeleteLineItemsFunc func(ctx context.Context, cartID int64) error
}

func (sm *StorageMock) AddProduct(ctx context.Context, cartID, productID int64, quantity uint32) error {
	return sm.AddProductFunc(ctx, cartID, productID, quantity)
}

func (sm *StorageMock) DeleteProduct(ctx context.Context, cartID, productID int64) error {
	return sm.DeleteProductFunc(ctx, cartID, productID)
}

func (sm *StorageMock) CartByID(ctx context.Context, id int64) (Cart, error) {
	return sm.CartByIDFunc(ctx, id)
}

func (sm *StorageMock) CreateCart(ctx context.Context, cart Cart) (Cart, error) {
	return sm.CreateCartFunc(ctx, cart)
}

func (sm *StorageMock) DeleteCart(ctx context.Context, cartID int64) error {
	return sm.DeleteCartFunc(ctx, cartID)
}

func (sm *StorageMock) DeleteLineItems(ctx context.Context, cartID int64) error {
	return sm.DeleteLineItemsFunc(ctx, cartID)
}
