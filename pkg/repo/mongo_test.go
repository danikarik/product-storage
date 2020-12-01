package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
			Prod: &Product{
				Name:      "Apple MacBook Pro",
				Price:     1299,
				UpdatedAt: now,
			},
		},
		{
			Prod: &Product{
				Name:      "Apple iPhone 12 PRO",
				Price:     1099,
				UpdatedAt: now,
			},
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
				tc.Prod.UpdatedAt = now.Add(1 * time.Hour)
				r.NoError(repo.SaveProduct(ctx, tc.Prod))

				new := repo.FindByName(ctx, tc.Prod.Name)
				r.NotNil(new)
				r.Len(new.Changes, 1)
				r.Equal(old.price, new.Changes[0])
				r.True(new.UpdatedAt.After(now))
			}
		})
	}
}
