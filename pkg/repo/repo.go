package repo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var errInvalidData = errors.New("repository: invalid input data")

type Product struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Price     float64            `bson:"price" json:"price"`
	Changes   []float64          `bson:"changes" json:"changes"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

type ProductOptions func(*Product)

func NewProduct(opts ...ProductOptions) *Product {
	p := &Product{ID: primitive.NewObjectID()}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type SortingDirection int

const (
	Desc SortingDirection = iota
	Asc
)

type SortingOption int

const (
	SortByDefault SortingOption = iota
	SortByName
	SortByPrice
	SortByUpdatedAt
)

type Pager struct {
	LastID primitive.ObjectID
	Limit  int64
}

type ListOptions struct {
	Direction SortingDirection
	Sorting   SortingOption
	Paging    *Pager
}

func (o *ListOptions) DirIndex() int {
	if o.Direction == Asc {
		return 1
	}
	return -1
}

// Repository holds methods to save and retrieve product information.
type Repository interface {
	FindByName(ctx context.Context, name string) *Product
	SaveProduct(ctx context.Context, p *Product) error
	ListProducts(ctx context.Context, opts *ListOptions) ([]Product, error)
}
