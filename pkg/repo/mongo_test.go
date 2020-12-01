package repo

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestSaveProduct(t *testing.T) {
	now := time.Now().UTC()

	testCases := []struct {
		Prod     *Product
		NewPrice float64
	}{
		{
			Prod: NewProduct(func(p *Product) {
				p.Name = "Apple MacBook Pro"
				p.Price = 1299
				p.UpdatedAt = now
			}),
		},
		{
			Prod: NewProduct(func(p *Product) {
				p.Name = "Apple iPhone 12 PRO"
				p.Price = 1099
				p.UpdatedAt = now
			}),
			NewPrice: 999,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Prod.Name, func(t *testing.T) {
			r := require.New(t)

			ctx := context.Background()

			conn, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
			r.NoError(err)
			r.NoError(conn.Connect(ctx))
			defer conn.Disconnect(ctx)

			r.NoError(conn.Database("productstore_test").Drop(ctx))

			repo := NewMongoRepo("productstore_test", conn)
			r.NoError(repo.SaveProduct(ctx, tc.Prod))

			if tc.NewPrice > 0 {
				old := struct {
					price     float64
					updatedAt time.Time
				}{
					price:     tc.Prod.Price,
					updatedAt: tc.Prod.UpdatedAt,
				}

				tc.Prod.Price = tc.NewPrice
				tc.Prod.UpdatedAt = tc.Prod.UpdatedAt.Add(1 * time.Hour)
				r.NoError(repo.SaveProduct(ctx, tc.Prod))

				new := repo.FindByName(ctx, tc.Prod.Name)
				r.NotNil(new)
				r.Len(new.Changes, 1)
				r.Equal(old.price, new.Changes[0])

				r.True(new.UpdatedAt.After(old.updatedAt))
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	randomID := primitive.NewObjectID()

	testCases := []struct {
		Name          string
		Products      []*Product
		Options       *ListOptions
		ExpectedCount int
	}{
		{
			Name: "DefaultSorting",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.Name = "Product1"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product2"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product3"
					p.Price = 2000
				}),
			},
			Options:       &ListOptions{},
			ExpectedCount: 3,
		},
		{
			Name: "SortByName",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.Name = "Apple"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Juice"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Banana"
					p.Price = 2000
				}),
			},
			Options:       &ListOptions{Sorting: SortByName, Direction: Asc},
			ExpectedCount: 3,
		},
		{
			Name: "SortByPrice",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.Name = "Apple"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Juice"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Banana"
					p.Price = 2000
				}),
			},
			Options:       &ListOptions{Sorting: SortByPrice, Direction: Desc},
			ExpectedCount: 3,
		},
		{
			Name: "SortByUpdatedAt",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.Name = "Apple"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Juice"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Banana"
					p.Price = 2000
				}),
			},
			Options:       &ListOptions{Sorting: SortByUpdatedAt, Direction: Desc},
			ExpectedCount: 3,
		},
		{
			Name: "Limit",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.Name = "Product1"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product2"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product3"
					p.Price = 2000
				}),
			},
			Options:       &ListOptions{Paging: &Pager{Limit: 1}},
			ExpectedCount: 1,
		},
		{
			Name: "Cursor",
			Products: []*Product{
				NewProduct(func(p *Product) {
					p.ID = randomID
					p.Name = "Product1"
					p.Price = 1000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product2"
					p.Price = 3000
				}),
				NewProduct(func(p *Product) {
					p.Name = "Product3"
					p.Price = 2000
				}),
			},
			Options: &ListOptions{Paging: &Pager{
				LastID: randomID,
				Limit:  1,
			}},
			ExpectedCount: 1,
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

			repo := NewMongoRepo("productstore_test", conn)

			now := time.Now().UTC()
			for i, prod := range tc.Products {
				prod.UpdatedAt = now.Add(time.Duration(i) * time.Hour)
				r.NoError(repo.SaveProduct(ctx, prod))
			}

			loaded, err := repo.ListProducts(ctx, tc.Options)
			r.NoError(err)
			r.Len(loaded, tc.ExpectedCount)

			testSortingOrder(t, loaded, tc.Options)
		})
	}
}

func testSortingOrder(t *testing.T, products []Product, opts *ListOptions) {
	r := require.New(t)

	switch opts.Sorting {
	case SortByName:
		sorted := sort.SliceIsSorted(products, func(i, j int) bool {
			if opts.Direction == Desc {
				return strings.Compare(products[i].Name, products[j].Name) == 1
			}
			return strings.Compare(products[i].Name, products[j].Name) == -1
		})
		r.True(sorted)
	case SortByPrice:
		sorted := sort.SliceIsSorted(products, func(i, j int) bool {
			if opts.Direction == Desc {
				return products[i].Price > products[j].Price
			}
			return products[i].Price < products[j].Price
		})
		r.True(sorted)
	case SortByUpdatedAt:
		sorted := sort.SliceIsSorted(products, func(i, j int) bool {
			if opts.Direction == Desc {
				return products[i].UpdatedAt.After(products[j].UpdatedAt)
			}

			return products[i].UpdatedAt.Before(products[j].UpdatedAt)
		})
		r.True(sorted)
	default:
		break
	}
}
