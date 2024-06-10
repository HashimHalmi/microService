package microServerMainFiles

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Product represents an item that can be purchased
type Product struct {
	ID          string  `bson:"id"`
	Name        string  `bson:"name"`
	Price       float64 `bson:"price"`
	Description string  `bson:"description"`
}

// Cart represents a shopping cart
type Cart struct {
	UserID    string     `bson:"user_id"`
	Items     []CartItem `bson:"items"`
	UpdatedAt time.Time  `bson:"updated_at"`
}

type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      string             `bson:"user_id"`
	Items       []CartItem         `bson:"items"`
	TotalAmount float64            `bson:"total_amount"`
	Status      string             `bson:"status"`
	CreatedAt   time.Time          `bson:"created_at"`
}
