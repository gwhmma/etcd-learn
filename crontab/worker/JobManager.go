package worker

import (
	"context"
	"encoding/json"
	"etcd-learn/crontab/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

//任务管理器
type EtcdManager struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

//定时任务
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
}

// 事件变化
type JobEvent struct {
	eventType int // save delete
	job       *Job
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
	watcher := clientv3.NewWatcher(client)

	Etcd = &EtcdManager{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	watchJobs()
}

//监听任务变化
func watchJobs() error {
	var jobEvent *JobEvent
	// get /cron/job/下的所有任务，并且获取当前集群的revision
	getResp, err := Etcd.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	//得到当前的所有任务
	for _, v := range getResp.Kvs {
		job := &Job{}
		if err := json.Unmarshal(v.Value, job); err != nil {
			continue
		}
		jobEvent = buildJobEvent(common.JOB_EVENT_SAVE, job)
		// 将这个任务同步给scheduler(调度协程)
		Schedule.PushJobEvent(jobEvent)
	}

	//从当前的revision向后监听事件变化
	// 监听协程
	go func() {
		// 从get时刻的后续版本开始监听变化
		watchStartRevision := getResp.Header.Revision + 1
		//监听/cron/job/目录下的变化
		watchChan := Etcd.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())

		//处理监听事件
		for watchResp := range watchChan {
			for _, w := range watchResp.Events {
				switch w.Type {
				case mvccpb.PUT: // 新建了任务 或是任务被修改
					job := &Job{}
					if err := json.Unmarshal(w.Kv.Value, job); err != nil {
						continue
					}

					//构建一个event事件
					jobEvent = buildJobEvent(common.JOB_EVENT_SAVE, job)

				case mvccpb.DELETE: // 删除了事件
					// 构建一个删除event
					job := &Job{Name: common.ExtractJobName(string(w.Kv.Key))}
					jobEvent = buildJobEvent(common.JOB_EVENT_DELETE, job)
					fmt.Println(*jobEvent)

				}

				//  推送一个事件给scheduler
				Schedule.PushJobEvent(jobEvent)
			}
		}

	}()
	return nil
}

//任务变化有2种 1. 更新任务 2.删除任务
func buildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		eventType: eventType,
		job:       job,
	}
}
