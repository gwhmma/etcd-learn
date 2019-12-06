package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

// 代表一个任务
type cronJob struct {
	expr *cronexpr.Expression
	next time.Time
}

func main() {
	// 需要一个调度协程检查所有cron任务 执行过期的任务
	scheduleJob := make(map[string]*cronJob)
	expr := cronexpr.MustParse("*/5 * * * * * *")
	now := time.Now()

	// 定义2个cronJob
	job1 := &cronJob{
		expr: expr,
		next: expr.Next(now),
	}

	job2 := &cronJob{
		expr: cronexpr.MustParse("*/6 * * * * * *"),
		next: expr.Next(now),
	}

	scheduleJob["job1"] = job1
	scheduleJob["job2"] = job2

	//启动一个调度协程
	go func() {
		//定时检查任务调度表
		for {
			now := time.Now()

			for jobName, job := range scheduleJob {
				if job.next.Before(now) || job.next.Equal(now) {
					//启动一个协程执行这个任务
					go func(jobName string) {
						fmt.Println("exec ", jobName)
					}(jobName)

					//计算下一次调度时间
					job.next = job.expr.Next(now)
					fmt.Println(jobName, "next exe time : ", job.next)
				}
			}

			//睡眠一段时间
			select {
			case <-time.NewTimer(time.Millisecond * 100).C: //将在100ms后可读 返回
			}
		}
	}()

	time.Sleep(100 * time.Second )
}
