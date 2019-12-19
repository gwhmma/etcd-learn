package common

import (
	"github.com/BurntSushi/toml"
	"strings"
)

type EtcdConfig struct {
	EtcdEndPoints   []string `toml:"etcdEndPoints"`
	EtcdDialTimeout int64    `toml:"etcdDialTimeout"`
}

type MongoConfig struct {
	MongoAddr []string `toml:"mongoAddr"`
	Timeout   int64  `toml:"timeout"`
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
