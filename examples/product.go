package examples

import (
	"time"
)

type Product struct {
	ID          string    `sql:"id,primary"`
	Name        string    `sql:"name"`
	Description string    `sql:"description"`
	Category    *string   `sql:"category"`
	Price       float64   `sql:"price"`
	Stock       int       `sql:"stock"`
	CreatedAt   time.Time `sql:"created_at"`
}

func (p *Product) TableName() string {
	return "products"
}
