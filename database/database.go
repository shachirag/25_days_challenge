	package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	ctx = context.Background()
)

type DB struct {
	client *mongo.Client
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func Connect() *DB {
	connectionString := os.Getenv("MONGO_URI")
	database := os.Getenv("DATABASE_NAME")

	if connectionString == "" || database == "" {
		log.Fatal("MONGO_URI and DATABASE_NAME must be set in .env file")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connectionString)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return &DB{
		client: client,
	}
}

func (db *DB) GetCollection(name string) *mongo.Collection {
	database := os.Getenv("DATABASE_NAME")
	return db.client.Database(database).Collection(name)
}

func (db *DB) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
