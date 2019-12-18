package master

import (
	"context"
	"encoding/json"
	"etcd-learn/crontab/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

//任务管理器
type EtcdManager struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

//定时任务
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
}

var Etcd *EtcdManager

//初始化etcd管理器
func InitEtcdManager(path string) {
	var (
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		etcdCfg *common.EtcdConfig
		err     error
	)

	if etcdCfg, err = common.LoadEtcdCfg(path); err != nil {
		fmt.Println("read etcdCfg err : ", err)
		return
	}

	//初始化配置
	config := clientv3.Config{
		Endpoints:   etcdCfg.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(etcdCfg.EtcdDialTimeout) * time.Millisecond, // 超时时间
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到kv和lease
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	Etcd = &EtcdManager{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
}

// 保存任务到etcd
// 把任务保存到 /con/jobs/任务名 目录下  --> json
func (j *Job) SaveJob() (oldJob *Job, err error) {
	var (
		jobValue []byte
		putResp  *clientv3.PutResponse
		old      Job
	)

	//etcd的保存key
	jobKey := fmt.Sprintf("%s%s", common.JOB_SAVE_DIR, j.Name)
	if jobValue, err = json.Marshal(j); err != nil {
		return
	}

	//保存到etcd
	if putResp, err = Etcd.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	//如果是更新 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		json.Unmarshal(putResp.PrevKv.Value, &old)
		oldJob = &old
	}
	return
}

//删除指定job
func (j *Job) DeleteJob() (oldJob *Job, err error) {
	var (
		old     Job
		delResp *clientv3.DeleteResponse
	)

	jobKey := fmt.Sprintf("%s%s", common.JOB_SAVE_DIR, j.Name)
	if delResp, err = Etcd.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	if len(delResp.PrevKvs) > 0 {
		json.Unmarshal(delResp.PrevKvs[0].Value, &old)
		oldJob = &old
	}
	return
}

//返回所有的任务
func (j *Job) JobList() ([]*Job, error) {
	var jobs []*Job

	getResp, err := Etcd.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil {
		return jobs, err
	}

	if len(getResp.Kvs) > 0 {
		for _, v := range getResp.Kvs {
			job := &Job{}
			if err := json.Unmarshal(v.Value, job); err != nil {
				continue
			}
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// kill一个任务
func (j *Job) KillJob() error {
	// 更新key = /cron/killer/任务名
	// 得到对应的key
	killKey := fmt.Sprintf("%s%s", common.JOB_KILL_DIR, j.Name)

	// 让worker监听到一次put操作，创建一个租约让其自动过期
	leaseGrant, err := Etcd.Lease.Grant(context.TODO(), 1)
	if err != nil {
		return err
	}

	if _, err := Etcd.Kv.Put(context.TODO(), killKey, "", clientv3.WithLease(leaseGrant.ID)); err != nil {
		return err
	}
	return nil
}
