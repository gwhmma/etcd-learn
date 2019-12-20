package common

import (
	"context"
	"github.com/BurntSushi/toml"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

type EtcdConfig struct {
	EtcdEndPoints   []string `toml:"etcdEndPoints"`
	EtcdDialTimeout int64    `toml:"etcdDialTimeout"`
}

type MongoConfig struct {
	MongoAddr []string `toml:"mongoAddr"`
	Timeout   int64    `toml:"timeout"`
}

// 任务执行日志过滤条件
type JobFilter struct {
	JobName string `bson:"jobName"`
}

// 任务执行日志排序条件  (-1)
type JobSort struct {
	Sort int64 `bson:"startTime"`
}

type Mongo struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func LoadEtcdCfg(path string) (*EtcdConfig, error) {
	etcd := &EtcdConfig{}
	if _, err := toml.DecodeFile(path, etcd); err != nil {
		return etcd, err
	}
	return etcd, nil
}

// /cron/job/job1  ---> job1
func ExtractJobName(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func LoadMongoCfg(path string) (*MongoConfig, error) {
	mongo := &MongoConfig{}
	if _, err := toml.DecodeFile(path, mongo); err != nil {
		return mongo, err
	}
	return mongo, nil
}

func MongoConn(path string) (*Mongo, error) {
	m := &Mongo{}

	mc, err := LoadMongoCfg(path)
	if err != nil {
		return m, err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(mc.Timeout)*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mc.MongoAddr[0]))
	if err != nil {
		return m, err
	}

	m.Client = client
	m.Collection = client.Database("cron").Collection("log")
	return m, nil
}
