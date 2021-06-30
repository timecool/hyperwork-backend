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
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb+srv://dev:yrua*!6_XQpvkW*@cluster0.7vsmw.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")

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
