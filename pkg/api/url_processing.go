package api

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/danikarik/product-storage/pkg/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) fetchData(ctx context.Context, url string) error {
	if url == "" {
		return status.Errorf(codes.InvalidArgument, "url must be specified")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "build request: %v", err)
	}

	resp, err := s.hclient.Do(req)
	if err != nil {
		return status.Errorf(codes.Internal, "send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return status.Errorf(codes.Internal, "wrong status: %d", resp.StatusCode)
	}

	if err := s.readCSV(ctx, resp.Body); err != nil {
		return status.Errorf(codes.Internal, "reading csv: %d", resp.StatusCode)
	}

	return nil
}

func (s *server) readCSV(ctx context.Context, r io.Reader) error {
	reader := csv.NewReader(r)
	reader.Comma = ';'

	// First read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Check format "PRODUCT NAME;PRICE"
	if len(header) != 2 || header[0] != "PRODUCT NAME" || header[1] != "PRICE" {
		return errors.New("invalid data format")
	}

	now := time.Now().UTC()

	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		price, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return fmt.Errorf("%s: %w", row[0], err)
		}

		prod := &repo.Product{
			Name:      row[0],
			Price:     price,
			UpdatedAt: now,
		}

		if err := s.repo.SaveProduct(ctx, prod); err != nil {
			return err
		}
	}

	return nil
}
