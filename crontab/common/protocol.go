package common

import (
	"github.com/BurntSushi/toml"
	"strings"
)

type EtcdConfig struct {
	EtcdEndPoints   []string `toml:"etcdEndPoints"`
	EtcdDialTimeout int64    `toml:"etcdDialTimeout"`
}

func LoadEtcdCfg(path string) (*EtcdConfig, error) {
	etcd := &EtcdConfig{}
	if _, err := toml.DecodeFile(path, etcd); err != nil {
		return etcd, err
	}
	return etcd, nil
}

// /cron/job/job1  ---> job1
func ExtractJobName(s string) string {
	return strings.TrimPrefix(s, JOB_SAVE_DIR)
}