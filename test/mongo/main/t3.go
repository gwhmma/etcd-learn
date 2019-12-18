package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

////记录任务的时间点
type TimePoint1 struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

type LogRecord1 struct {
	JobName   string    `bson:"jobName"` //任务名
	Command   string    `bson:"command"` // shell命令
	Err       string    `bson:"err"`     // 报错信息
	Content   string    `bson:"content"` // 脚本输出内容
	TimePoint TimePoint1 `bson:"timePoint"`
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		res        *mongo.InsertManyResult
		err        error
	)

	//1 建立连接
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017")); err != nil {
		fmt.Println(err, " 1")
		return
	}

	//2 选择数据库 my_db
	database = client.Database("crontab")
	//3 选择表 gwh
	collection = database.Collection("log")
	//4  bson
	record := &LogRecord1{
		JobName: "job10",
		Command: "echo hello",
		Err:     "",
		Content: "hello",
		TimePoint: TimePoint1{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}

	// 批量插入多条
	logArr := []interface{}{record, record, record}
	if res, err = collection.InsertMany(context.TODO(), logArr); err != nil {
		fmt.Println(err)
		return
	}

	// 分布式集群下的ID生成算法
	// snowflake : 毫秒/微秒的当前时间 + 机器ID + 毫秒/微秒内的自增ID(每当毫秒变化了，会重置为0， 继续自增)
	for _, id := range res.InsertedIDs {
		fmt.Println(id.(primitive.ObjectID).Hex())
	}

}
