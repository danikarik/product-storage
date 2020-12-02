package api

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/danikarik/product-storage/pkg/repo"
	"github.com/danikarik/product-storage/pkg/store"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestServerFetch(t *testing.T) {
	ctx := context.Background()

	conn, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	require.NoError(t, conn.Connect(ctx))
	defer conn.Disconnect(ctx)

	srv := &server{
		repo:    repo.NewMongoRepo("productstore_test", conn),
		timeout: _defaultTimeout,
		hclient: &http.Client{},
	}

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
			r.NoError(conn.Database("productstore_test").Drop(ctx))

			resp, err := srv.Fetch(ctx, &store.FetchRequest{Url: tc.URL})
			r.NoError(err)
			r.Equal(int32(0), resp.Result)
		})
	}
}

func TestServerList(t *testing.T) {
	ctx := context.Background()

	conn, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	require.NoError(t, conn.Connect(ctx))
	defer conn.Disconnect(ctx)

	srv := &server{
		repo:    repo.NewMongoRepo("productstore_test", conn),
		timeout: _defaultTimeout,
		hclient: &http.Client{},
	}

	testCases := []struct {
		Name          string
		URL           string
		Paging        *store.Paging
		Sorting       *store.Sorting
		ExpectedCount int
	}{
		{
			Name:          "Dummy",
			URL:           "https://csv-samples.s3.amazonaws.com/dummy.csv",
			Paging:        &store.Paging{},
			Sorting:       &store.Sorting{},
			ExpectedCount: 9,
		},
		{
			Name:          "Phones",
			URL:           "https://csv-samples.s3.amazonaws.com/iphones.csv",
			Paging:        &store.Paging{},
			Sorting:       &store.Sorting{},
			ExpectedCount: 2,
		},
		{
			Name:          "Limit",
			URL:           "https://csv-samples.s3.amazonaws.com/dummy.csv",
			Paging:        &store.Paging{Limit: 5},
			Sorting:       &store.Sorting{},
			ExpectedCount: 5,
		},
		{
			Name:   "SortingByNameDesc",
			URL:    "https://csv-samples.s3.amazonaws.com/dummy.csv",
			Paging: &store.Paging{},
			Sorting: &store.Sorting{
				Direction: store.Direction_DESC,
				Field:     store.Field_NAME,
			},
			ExpectedCount: 9,
		},
		{
			Name:   "SortingByPriceAsc",
			URL:    "https://csv-samples.s3.amazonaws.com/dummy.csv",
			Paging: &store.Paging{},
			Sorting: &store.Sorting{
				Direction: store.Direction_ASC,
				Field:     store.Field_PRICE,
			},
			ExpectedCount: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)
			r.NoError(conn.Database("productstore_test").Drop(ctx))

			resp, err := srv.Fetch(ctx, &store.FetchRequest{Url: tc.URL})
			r.NoError(err)
			r.Equal(int32(0), resp.Result)

			result, err := srv.List(ctx, &store.ListRequest{
				Paging:  tc.Paging,
				Sorting: tc.Sorting,
			})

			r.NoError(err)
			r.Len(result.Products, tc.ExpectedCount)

			if tc.Paging.Limit > 0 {
				r.NotEmpty(result.LastId)

				// try to fetch second page
				result, err = srv.List(ctx, &store.ListRequest{
					Paging: &store.Paging{
						LastId: result.LastId,
						Limit:  tc.Paging.Limit,
					},
					Sorting: &store.Sorting{},
				})

				r.NoError(err)
			}

			if tc.Sorting.Field != store.Field_DEFAULT {
				testSortingOrder(t, result.Products, tc.Sorting)
			}
		})
	}
}

func testSortingOrder(t *testing.T, products []*store.Product, opts *store.Sorting) {
	r := require.New(t)

	switch opts.Field {
	case store.Field_NAME:
		sorted := sort.SliceIsSorted(products, func(i, j int) bool {
			if opts.Direction == store.Direction_DESC {
				return strings.Compare(products[i].Name, products[j].Name) == 1
			}
			return strings.Compare(products[i].Name, products[j].Name) == -1
		})
		r.True(sorted)
	case store.Field_PRICE:
		sorted := sort.SliceIsSorted(products, func(i, j int) bool {
			if opts.Direction == store.Direction_DESC {
				return products[i].Price > products[j].Price
			}
			return products[i].Price < products[j].Price
		})
		r.True(sorted)
	default:
		break
	}
}
