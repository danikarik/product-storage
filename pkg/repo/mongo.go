package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepo struct {
	name   string
	client *mongo.Client
}

func NewMongoRepo(name string, client *mongo.Client) Repository {
	return &mongoRepo{name: name, client: client}
}

func (m *mongoRepo) db() *mongo.Database {
	return m.client.Database(m.name)
}

func (m *mongoRepo) FindByName(ctx context.Context, name string) *Product {
	var p Product
	if err := m.db().Collection("products").FindOne(ctx, bson.M{"name": name}).Decode(&p); err == nil {
		return &p
	}

	return nil
}

func (m *mongoRepo) SaveProduct(ctx context.Context, p *Product) error {
	if p == nil {
		return errInvalidData
	}

	old := m.FindByName(ctx, p.Name)
	if old == nil {
		// insert new record
		if _, err := m.db().Collection("products").InsertOne(ctx, p); err != nil {
			return err
		}

		return nil
	}

	// return if no changes
	if old.Price == p.Price {
		return nil
	}

	var (
		filter = bson.M{"name": p.Name}
		update = bson.M{
			"$set": bson.M{
				"price":     p.Price,
				"updatedat": p.UpdatedAt,
				"changes":   append(old.Changes, old.Price),
			},
		}
	)

	if _, err := m.db().Collection("products").UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}
