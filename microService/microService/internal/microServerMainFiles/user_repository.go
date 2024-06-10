package microServerMainFiles

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var client *mongo.Client
var usersCollection *mongo.Collection

func init() {
	// Connect to MongoDB
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	// Get the users collection
	usersCollection = client.Database("authDB").Collection("users")
}

func SaveUser(user UserCredentials) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := usersCollection.InsertOne(ctx, user)
	return err
}

func GetUserByEmail(email string) (UserCredentials, error) {
	var user UserCredentials
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return user, err
}
