package worker

import (
	"context"
	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {
}

// 任务执行结果
type JobExecuteResult struct {
	exeInfo   *JobExecuteInfo //执行任务状态
	output    []byte          // 执行输出
	err       error           //执行错误信息
	startTime time.Time       //任务启动时间
	endTime   time.Time       // 任务执行完成时间
}

var Exe *Executor

//初始化执行器
func InitExecutor() error {
	Exe = &Executor{}
	return nil
}

// 执行一个任务
func (e *Executor) executeJob(exeInfo *JobExecuteInfo) {
	go func() {
		exeRes := &JobExecuteResult{
			exeInfo: exeInfo,
			output:  make([]byte, 0),
		}

		// 首先获取分布式锁
		// 初始化分布式锁
		start := time.Now()
		jobLock := Etcd.createJobLock(exeInfo.Job.Name)
		err := jobLock.tryLock()
		defer jobLock.unlock()

		if err != nil {
			exeRes.startTime = start
			exeRes.endTime = time.Now()
			exeRes.err = err
		} else {
			//执行shell命令
			exeRes.startTime = start
			cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", exeInfo.Job.Command)
			//执行命令并捕获错误
			res, err := cmd.CombinedOutput()
			exeRes.output = res
			exeRes.err = err
			exeRes.endTime = time.Now()
		}

		//任务执行完成后 将执行结果返回给scheduler 并把该条记录从任务列表中删除
		Schedule.pushJobExeRes(exeRes)
	}()
}
