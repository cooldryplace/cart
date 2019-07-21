package cart

import (
	"testing"
	"time"

	"github.com/cooldryplace/proto"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
)

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
			name:  "Empty values",
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
