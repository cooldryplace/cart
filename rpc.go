package cart

import (
	"context"
	"fmt"

	"github.com/cooldryplace/proto"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var emptyResp = &empty.Empty{}

func toProtoLineItem(li LineItem) *proto.LineItem {
	return &proto.LineItem{
		ProductId: li.ProductID,
		Quantity:  li.Quantity,
	}
}

func toProtoLineItems(lis []LineItem) []*proto.LineItem {
	result := make([]*proto.LineItem, 0, len(lis))

	for _, li := range lis {
		result = append(result, toProtoLineItem(li))
	}

	return result
}

func toProtoCart(c Cart) (*proto.Cart, error) {
	createdAt, err := ptypes.TimestampProto(c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert creation time to ptypes.Timestamp: %s", err)
	}
	updatedAt, err := ptypes.TimestampProto(c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert update time to ptypes.Timestamp: %s", err)
	}

	return &proto.Cart{
		Id:        c.ID,
		UserId:    c.UserID,
		Items:     toProtoLineItems(c.Items),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// Server implements protobuf Carts service.
type Server struct {
	carts *Carts
}

// NewServer returns Carts gRPC server.
func NewServer(c *Carts) *Server {
	return &Server{carts: c}
}

// AddProduct to a Cart.
func (s *Server) AddProduct(ctx context.Context, req *proto.AddProductRequest) (*empty.Empty, error) {
	if req.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "failed to add the product: wrong quantity")
	}

	if err := s.carts.AddProduct(ctx, req.CartId, req.ProductId, req.Quantity); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add the Product: %s", err)
	}

	return emptyResp, nil
}

// DelProduct removes product from a Cart.
func (s *Server) DelProduct(ctx context.Context, req *proto.DelProductRequest) (*empty.Empty, error) {
	if err := s.carts.DeleteProduct(ctx, req.CartId, req.ProductId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete the Product: %s", err)
	}

	return emptyResp, nil
}

// CreateCart for a User.
func (s *Server) CreateCart(ctx context.Context, req *proto.CartCreateRequest) (*proto.CartResponse, error) {
	cart, err := s.carts.Create(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create the Cart: %s", err)
	}

	pCart, err := toProtoCart(cart)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert the Cart: %s", err)
	}

	return &proto.CartResponse{Cart: pCart}, nil
}

// DeleteCart with the matching ID.
func (s *Server) DeleteCart(ctx context.Context, req *proto.CartDeleteRequest) (*empty.Empty, error) {
	if err := s.carts.Delete(ctx, req.Id); err != nil {
		if err == errNotFound {
			return nil, status.Errorf(codes.NotFound, "cart with ID: %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete the Cart: %s", err)
	}

	return emptyResp, nil
}

// EmptyCart deletes all LineItems from a Cart.
func (s *Server) EmptyCart(ctx context.Context, req *proto.EmptyCartRequest) (*empty.Empty, error) {
	if err := s.carts.Empty(ctx, req.CartId); err != nil {
		if err == errNotFound {
			return nil, status.Errorf(codes.NotFound, "cart with ID: %d not found", req.CartId)
		}
		return nil, status.Errorf(codes.Internal, "failed to empty the Cart: %s", err)
	}

	return emptyResp, nil
}

// GetCart returns current Cart state.
func (s *Server) GetCart(ctx context.Context, req *proto.CartRequest) (*proto.CartResponse, error) {
	cart, err := s.carts.Cart(ctx, req.Id)
	if err != nil {
		if err == errNotFound {
			return nil, status.Errorf(codes.NotFound, "cart with ID: %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get the Cart: %s", err)
	}

	pCart, err := toProtoCart(cart)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert the Cart: %s", err)
	}

	return &proto.CartResponse{Cart: pCart}, nil
}
