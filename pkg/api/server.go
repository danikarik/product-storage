package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danikarik/product-storage/pkg/repo"
	pb "github.com/danikarik/product-storage/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const _defaultTimeout = 30 * time.Second

type server struct {
	pb.UnimplementedStoreServer

	repo    repo.Repository
	timeout time.Duration
	hclient *http.Client
}

// Options holds server configuration.
type Options struct{ Timeout time.Duration }

// NewServer returns server stub with implemented methods.
func NewServer(repo repo.Repository, opts *Options) *grpc.Server {
	if opts == nil {
		opts = &Options{Timeout: _defaultTimeout}
	}

	srv := &server{
		repo:    repo,
		timeout: opts.Timeout,
		hclient: &http.Client{},
	}

	s := grpc.NewServer()
	pb.RegisterStoreServer(s, srv)
	return s
}

func (s *server) Fetch(ctx context.Context, in *pb.FetchRequest) (*pb.FetchResponse, error) {
	if err := s.fetchData(ctx, in.Url); err != nil {
		return nil, err
	}

	return &pb.FetchResponse{Result: 0}, nil
}

func (s *server) List(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	if in.Paging == nil || in.Sorting == nil {
		return nil, status.Errorf(codes.InvalidArgument, "url must be specified")
	}

	opts, err := buildListOptions(in)
	if err != nil {
		return nil, err
	}

	products, err := s.repo.ListProducts(ctx, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not retrieve products")
	}

	resp := &pb.ListResponse{
		LastId:   "",
		Products: make([]*pb.Product, 0),
	}

	if len(products) > 0 {
		resp.LastId = products[len(products)-1].ID.Hex()
	}

	for _, p := range products {
		resp.Products = append(resp.Products, &pb.Product{
			Name:         p.Name,
			Price:        p.Price,
			NumOfChanges: int64(len(p.Changes)),
			LastUpdate:   p.UpdatedAt.String(),
		})
	}

	return resp, nil
}

func buildListOptions(in *pb.ListRequest) (*repo.ListOptions, error) {
	opts := &repo.ListOptions{}

	opts.Direction = repo.Desc
	if in.Sorting.Direction == pb.Direction_ASC {
		opts.Direction = repo.Asc
	}

	switch in.Sorting.Field {
	case pb.Field_NAME:
		opts.Sorting = repo.SortByName
	case pb.Field_PRICE:
		opts.Sorting = repo.SortByPrice
	case pb.Field_UPDATED:
		opts.Sorting = repo.SortByUpdatedAt
	default:
		opts.Sorting = repo.SortByDefault
	}

	opts.Paging = &repo.Pager{Limit: in.Paging.Limit}
	if in.Paging.LastId != "" {
		id, err := primitive.ObjectIDFromHex(in.Paging.LastId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid object id")
		}

		opts.Paging.LastID = id
	}

	return opts, nil
}
