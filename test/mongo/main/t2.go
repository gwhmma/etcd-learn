package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//记录任务的时间点
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

type LogRecord struct {
	JobName   string    `bson:"jobName"` //任务名
	Command   string    `bson:"command"` // shell命令
	Err       string    `bson:"err"`     // 报错信息
	Content   string    `bson:"content"` // 脚本输出内容
	TimePoint TimePoint `bson:"timePoint"`
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		res        *mongo.InsertOneResult
		docID      primitive.ObjectID
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
	//4  bson
	record := &LogRecord{
		JobName: "job10",
		Command: "echo hello",
		Err:     "",
		Content: "hello",
		TimePoint: TimePoint{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}

	if res, err = collection.InsertOne(context.TODO(), record); err != nil {
		fmt.Println(err, " 2 ")
		return
	}

	// _id : 默认生成一个全局唯一id， 12字节的二进制
	docID = res.InsertedID.(primitive.ObjectID)
	fmt.Println("自增id ", docID.Hex())
}
