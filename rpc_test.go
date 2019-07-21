package cart

import (
	"testing"
	"time"

	"github.com/cooldryplace/proto"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
)

func TestToProtoLineItems(t *testing.T) {
	cases := []struct {
		name     string
		input    LineItem
		expected *proto.LineItem
	}{
		{
			name:     "Zero values",
			input:    LineItem{},
			expected: &proto.LineItem{},
		},
		{
			name: "All values",
			input: LineItem{
				ProductID: 100500,
				Quantity:  1,
			},
			expected: &proto.LineItem{
				ProductId: 100500,
				Quantity:  1,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := toProtoLineItem(c.input)

			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("toProtoLineItem() mismatch (+got -want)\n%s", diff)
			}
		})
	}
}

func TestToProtoCart(t *testing.T) {
	var (
		lineItem       = LineItem{ProductID: 13, Quantity: 100}
		lineItems      = []LineItem{lineItem}
		createTime     = time.Now()
		updateTime     = createTime.Add(1 * time.Second)
		pZeroTime, _   = ptypes.TimestampProto(time.Time{})
		pCreateTime, _ = ptypes.TimestampProto(createTime)
		pUpdateTime, _ = ptypes.TimestampProto(updateTime)
	)

	cases := []struct {
		name          string
		input         Cart
		expected      *proto.Cart
		expectedError error
	}{
		{
			name:  "Zero values",
			input: Cart{},
			expected: &proto.Cart{
				Items:     []*proto.LineItem{},
				CreatedAt: pZeroTime,
				UpdatedAt: pZeroTime,
			},
			expectedError: nil,
		},
		{
			name: "All values",
			input: Cart{
				ID:        100500,
				UserID:    42,
				Items:     lineItems,
				CreatedAt: createTime,
				UpdatedAt: updateTime,
			},
			expected: &proto.Cart{
				Id:        100500,
				UserId:    42,
				Items:     toProtoLineItems(lineItems),
				CreatedAt: pCreateTime,
				UpdatedAt: pUpdateTime,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := toProtoCart(c.input)
			if err != c.expectedError {
				t.Fatalf("Got error: %v, expected: %v", err, c.expectedError)
			}

			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("toProtoCart() mismatch (+got -want)\n%s", diff)
			}
		})
	}
}
