package master

import (
	"context"
	"etcd-learn/crontab/common"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type WorkerManager struct {
	Client *clientv3.Client
	Kv     clientv3.KV
}

var Wm *WorkerManager

func InitWorkerManager(path string) error {
	etcdCfg, err := common.LoadEtcdCfg(path)
	if err != nil {
		return err
	}

	config := clientv3.Config{
		Endpoints:   etcdCfg.EtcdEndPoints,
		DialTimeout: time.Duration(etcdCfg.EtcdDialTimeout) * time.Millisecond,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	Wm = &WorkerManager{
		Client: client,
		Kv:     client.KV,
	}
	return nil
}

func WorkerList() ([]string, error) {
	list := make([]string, 0)

	getResp, err := Wm.Kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix())
	if err != nil {
		return list, err
	}

	for _, v := range getResp.Kvs {
		list = append(list, common.ExtractJobName(string(v.Key), common.JOB_WORKER_DIR))
	}

	return list, nil
}
