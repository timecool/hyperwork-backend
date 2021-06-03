package database

import (
	"context"
	"fmt"
	"log"
	"time"
	"timecool/hyperwork/util"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Globle Variable for Database
var DB *mongo.Database
var Ctx = context.TODO()

func Connect() {
	// Set Username and Password
	credential := options.Credential{
		Username: util.GetEnvVariable("MONGODB_USERNAME"),
		Password: util.GetEnvVariable("MONGODB_PASSWORD"),
	}
	applyUri := fmt.Sprintf("mongodb://%s:%s/", util.GetEnvVariable("MONGODB_HOST"), util.GetEnvVariable("MONGODB_PORT"))
	// Set client options
	clientOptions := options.Client().ApplyURI(applyUri).SetAuth(credential)

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
