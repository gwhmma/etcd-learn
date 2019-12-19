package worker

import (
	"context"
	"etcd-learn/crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//MongoDB存储日志
type LogSink struct {
	client         *mongo.Client
	collection     *mongo.Collection
	logChan        chan *JobLog
	autoCommitChan chan *LogBatch
}

type LogBatch struct {
	logs []interface{}
}

var Sink *LogSink

func InitLogSink(path string) error {
	//连接到MongoDB
	mc, err := common.LoadMongoCfg(path)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mc.MongoAddr[0]))
	if err != nil {
		return err
	}

	Sink = &LogSink{
		client:         client,
		collection:     client.Database("cron").Collection("log"),
		logChan:        make(chan *JobLog, 1000),
		autoCommitChan: make(chan *LogBatch, 1000),
	}

	go Sink.writeLoop()

	return nil
}

//日志存储协程
func (s *LogSink) writeLoop() {
	var batch *LogBatch
	var commitTimer *time.Timer

	for {
		select {
		case log := <-s.logChan:
			//写入mongodb
			// 每次写入需要请求一次MongoDB 耗时(网络缓慢)  所以按批次插入
			if batch == nil {
				batch = &LogBatch{}
				// 让这个批次超时自动提交 (比如1秒)
				commitTimer = time.AfterFunc(time.Second*1,
					func(batch *LogBatch) func() {
						// 发出超时通知 不直接提交batch
						return func() {
							Sink.autoCommitChan <- batch
						}
					}(batch))
			}

			batch.logs = append(batch.logs, log)
			if len(batch.logs) >= 100 {
				s.collection.InsertMany(context.TODO(), batch.logs)
				batch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch := <-s.autoCommitChan: // 过期的批次
			// 把批次写入到MongoDB
			// 判断超时批次是否为当前批次
			if timeoutBatch != batch {
				// 跳过已提交的批次
				continue
			}
			s.collection.InsertMany(context.TODO(), timeoutBatch.logs)
			batch = nil
		}
	}
}

//发送日志
func (s *LogSink) Append(log *JobLog) {
	select {
	case s.logChan <- log:
	default:
		// 队列满了就丢弃
	}
}
