package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
				"price":      p.Price,
				"updated_at": p.UpdatedAt,
				"changes":    append(old.Changes, old.Price),
			},
		}
	)

	if _, err := m.db().Collection("products").UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (m *mongoRepo) ListProducts(ctx context.Context, opts *ListOptions) ([]Product, error) {
	buildListOptions(opts)

	filter := bson.M{}
	if len(opts.Paging.LastID) > 0 {
		filter = bson.M{"_id": bson.M{"$gt": opts.Paging.LastID}}
	}

	cursor, err := m.db().Collection("products").Find(ctx, filter, buildFindOptions(opts))
	if err != nil {
		return nil, err
	}

	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func buildListOptions(opts *ListOptions) {
	if opts == nil {
		opts = &ListOptions{
			Direction: Desc,
			Sorting:   SortByDefault,
			Paging:    &Pager{Limit: 10},
		}
	}

	if opts.Paging == nil {
		opts.Paging = &Pager{Limit: 10}
	} else {
		if opts.Paging.Limit == 0 {
			opts.Paging.Limit = 10
		}
	}
}

func buildFindOptions(opts *ListOptions) *options.FindOptions {
	fopts := options.Find()

	switch opts.Sorting {
	case SortByName:
		fopts.SetSort(bson.M{"name": opts.DirIndex()})
	case SortByPrice:
		fopts.SetSort(bson.M{"price": opts.DirIndex()})
	case SortByUpdatedAt:
		fopts.SetSort(bson.M{"updated_at": opts.DirIndex()})
	default:
		break
	}

	fopts.SetLimit(opts.Paging.Limit)

	return fopts
}
