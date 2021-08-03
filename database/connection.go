package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

//Globle Variable for Database
var DB *mongo.Database
var Ctx = context.TODO()

func Connect() {
	// Set Username and Password
	credential := options.Credential{
		Username: "dev",
		Password: "dev",
	}
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/").SetAuth(credential)

	Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	//Set Database
	DB = client.Database("officeDb")

}
