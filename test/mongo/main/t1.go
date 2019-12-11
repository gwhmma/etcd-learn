package main

import (
	"context"
	"fmt"
	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func main() {
	var (
		client *mongo.Client
		database *mongo.Database
		collection *mongo.Collection
		err error
	)

	//建立连接
	//mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017", clientopt.ConnectTimeout)
	t := time.Second * 5
	options := &options.ClientOptions{
		Hosts: []string{"mongodb://127.0.0.1:27017"},
		ConnectTimeout: &t,
	}
	if client, err = mongo.Connect(context.TODO(), options); err != nil {
		fmt.Println(err)
		return
	}

	//选择数据库 my_db
	database = client.Database("my_db")

	//选择表 gwh
	collection = database.Collection("gwh")

	collection = collection
}
