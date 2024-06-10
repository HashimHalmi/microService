package microServerMainFiles

import "go.mongodb.org/mongo-driver/mongo"

var db *mongo.Database

func SetDatabase(database *mongo.Database) {
	db = database
}
