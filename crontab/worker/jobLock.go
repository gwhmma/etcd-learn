package worker

import (
	"context"
	"etcd-learn/crontab/common"
	"github.com/coreos/etcd/clientv3"
)

//分布式锁 (txn事务)
type jobLock struct {
	// etcd客户端
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string             //任务名
	cancelFunc context.CancelFunc // 用于自动续租
	leaseId    clientv3.LeaseID   //租约id
	lock       bool
}

//初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) *jobLock {
	return &jobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
}

//尝试上锁
func (j *jobLock) tryLock() error {
	// 1. 创建租约 5秒
	leaseGrantResp, err := j.lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}

	// 2. 创建ctx 用于取消自动续租
	ctx, cancelFunc := context.WithCancel(context.TODO())

	// 3. 自动续租
	leaseId := leaseGrantResp.ID
	leaseAliveChan, err := j.lease.KeepAlive(ctx, leaseId)
	if err != nil {
		// 取消自动续租
		cancelFunc()
		// 释放租约
		j.lease.Revoke(context.TODO(), leaseId)
		return err
	}

	//4 .处理自动续租的协程
	go func() {
		for {
			select {
			case keepRsp := <-leaseAliveChan:
				if keepRsp == nil {
					goto END
				}
			}
		}
	END:
	}()

	// 5. 创建txn事务
	txn := j.kv.Txn(context.TODO())
	lockKey := common.JOB_LOCK_DIR + j.jobName

	// 6. 事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).Else(clientv3.OpGet(lockKey))

	// 提交事务
	txnResp, err := txn.Commit()
	if err != nil {
		cancelFunc()
		j.lease.Revoke(context.TODO(), leaseId)
		return err
	}

	// 7. 成功返回, 抢锁失败释放租约
	if !txnResp.Succeeded {
		//锁被占用
		cancelFunc()
		j.lease.Revoke(context.TODO(), leaseId)
		return common.ERR_LOCK_ALADY_REQUIRED
	}

	//抢锁成功
	j.cancelFunc = cancelFunc
	j.leaseId = leaseId
	j.lock = true

	return nil
}

// 释放锁
func (j *jobLock) unlock() {
	if j.lock {
		j.cancelFunc()
		j.lease.Revoke(context.TODO(), j.leaseId)
	}
}
