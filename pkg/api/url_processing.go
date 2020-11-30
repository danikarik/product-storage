package api

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"

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

	if err := s.readCSV(resp.Body); err != nil {
		return status.Errorf(codes.Internal, "reading csv: %d", resp.StatusCode)
	}

	return nil
}

func (s *server) readCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.Comma = ';'

	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		log.Println(row)
	}

	return nil
}
