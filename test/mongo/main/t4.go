package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

////记录任务的时间点
type TimePoint2 struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

type LogRecord2 struct {
	JobName   string     `bson:"jobName"` //任务名
	Command   string     `bson:"command"` // shell命令
	Err       string     `bson:"err"`     // 报错信息
	Content   string     `bson:"content"` // 脚本输出内容
	TimePoint TimePoint2 `bson:"timePoint"`
}

type FindByJobName struct {
	JobName string `bson:"jobName"`
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		cur        *mongo.Cursor
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
	//4  bson   按照jobName过滤, 找出jobName=job10, 找出5条
	condition := &FindByJobName{JobName: "job10",} // {"jobName":"job10"}

	//查询
	skip := int64(0)
	limit := int64(2)

	defer cur.Close(context.TODO())

	if cur, err = collection.Find(context.TODO(), condition, &options.FindOptions{Skip: &skip, Limit: &limit,}); err != nil {
		fmt.Println(err)
		return
	}

	//遍历结果
	for cur.Next(context.TODO()) {
		record := &LogRecord2{}
		cur.Decode(record)
		fmt.Println(record)
	}

}
