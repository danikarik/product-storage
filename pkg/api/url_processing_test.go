package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/danikarik/product-storage/pkg/repo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestFetchData(t *testing.T) {
	testCases := []struct {
		Name string
		URL  string
	}{
		{
			Name: "Dummy",
			URL:  "https://csv-samples.s3.amazonaws.com/dummy.csv",
		},
		{
			Name: "Phones",
			URL:  "https://csv-samples.s3.amazonaws.com/iphones.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			ctx := context.Background()

			conn, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
			r.NoError(err)
			r.NoError(conn.Connect(ctx))
			defer conn.Disconnect(ctx)

			r.NoError(conn.Database("productstore_test").Drop(ctx))

			s := &server{
				timeout: _defaultTimeout,
				hclient: &http.Client{},
				repo:    repo.NewMongoRepo("productstore_test", conn),
			}

			r.NoError(s.fetchData(context.Background(), tc.URL))
		})
	}
}

func TestReadCSV(t *testing.T) {
	testCases := []struct {
		Name string
		Path string
	}{
		{
			Name: "Dummy",
			Path: "testdata/dummy.csv",
		},
		{
			Name: "Phones",
			Path: "testdata/iphones.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			ctx := context.Background()

			conn, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
			r.NoError(err)
			r.NoError(conn.Connect(ctx))
			defer conn.Disconnect(ctx)

			r.NoError(conn.Database("productstore_test").Drop(ctx))

			s := &server{
				timeout: _defaultTimeout,
				hclient: &http.Client{},
				repo:    repo.NewMongoRepo("productstore_test", conn),
			}

			file, err := os.Open(tc.Path)
			r.NoError(err)
			defer file.Close()

			r.NoError(s.readCSV(context.Background(), file))
		})
	}
}
