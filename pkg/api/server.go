package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danikarik/product-storage/pkg/repo"
	pb "github.com/danikarik/product-storage/pkg/store"
	"google.golang.org/grpc"
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
	if err := s.fetchData(ctx, in.GetUrl()); err != nil {
		return nil, err
	}

	return &pb.FetchResponse{Result: 0}, nil
}
