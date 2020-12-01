package repo

import (
	"context"
	"errors"
	"time"
)

var errInvalidData = errors.New("repository: invalid input data")

type Product struct {
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Changes   []float64 `json:"changes"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Repository holds methods to save and retrieve product information.
type Repository interface {
	FindByName(ctx context.Context, name string) *Product
	SaveProduct(ctx context.Context, p *Product) error
}
