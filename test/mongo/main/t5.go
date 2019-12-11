package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// {"$lt":timestamp}
type TimeBeforeCond struct {
	Before int64 `bson:"$lt"`
}

// {"timePoint.startTime":{"$lt":timestamp} }
type DeleteCond struct {
	BeforeCond TimeBeforeCond `bson:"timePoint.startTime"`
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		del        *mongo.DeleteResult
		err        error
	)

	//1 建立连接
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017")); err != nil {
		fmt.Println(err, " 1")
		return
	}

	//2 选择数据库 my_db
	database = client.Database("cron")
	//3 选择表 gwh
	collection = database.Collection("log")

	// 删除开始时间早于当前时间的记录
	// delete{"timePoint.startTime":{"$lt":timestamp} }
	delCond := &DeleteCond{BeforeCond: TimeBeforeCond{Before:time.Now().Unix()}}
	if del, err = collection.DeleteMany(context.TODO(), delCond); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(del.DeletedCount)
}
