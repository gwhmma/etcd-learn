package worker

import (
	"context"
	"etcd-learn/crontab/common"
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

//任务调度
type Scheduler struct {
	JobEventChan    chan *JobEvent              //etcd任务事件队列
	JobPlanMap      map[string]*JobSchedulePlan // 任务调度计划列表
	JobExecutingMap map[string]*JobExecuteInfo  // 任务的执行状态列表
	JobExeResChan   chan *JobExecuteResult      // 任务执行结果chan
}

//任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 调度的任务信息
	Expr     *cronexpr.Expression // 解析好的 cronexpr 表达式
	NextTime time.Time
}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               //任务信息
	PlanTime   time.Time          //计划执行时间
	RealTime   time.Time          //实际执行时间
	Ctx        context.Context    // 取消command执行的context
	CancelFunc context.CancelFunc //取消command的cancel函数
}

type JobLog struct {
	JobName      string `bson:"jobName" json:"jobName"`           //任务名
	Command      string `bson:"command" json:"command"`           // 执行的命令
	OutPut       string `bson:"outPut" json:"outPut"`             // 执行结果输出
	Error        string `bson:"error" json:"error"`               //报错原因
	PlanTime     int64  `bson:"planTime" json:"planTime"`         // 计划开始时间
	ScheduleTime int64  `bson:"scheduleTime" json:"scheduleTime"` //任务调度时间
	StartTime    int64  `bson:"startTime" json:"startTime"`       // 任务开始执行时间
	EndTime      int64  `bson:"endTime" json:"endTime"`           // 任务结束时间
}

var Schedule *Scheduler

//初始化调度器
func InitScheduler() {
	Schedule = &Scheduler{
		JobEventChan:    make(chan *JobEvent, 1000),
		JobPlanMap:      make(map[string]*JobSchedulePlan),
		JobExecutingMap: make(map[string]*JobExecuteInfo),
		JobExeResChan:   make(chan *JobExecuteResult, 1000),
	}

	go Schedule.scheduleLoop()
}

//调度协程
func (s *Scheduler) scheduleLoop() {
	//初始化一次 (第一次是一秒)
	scheduleAfter := s.trySchedule()

	// 调度的延时定时器
	timer := time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent := <-s.JobEventChan: // 监听任务变化事件
			// 对内存中维护的任务列表进行增删改查
			s.HandleJobEvent(jobEvent)
		case <-timer.C: // 最近的任务到期了
		case jobExeRes := <-s.JobExeResChan: //监听任务执行结果
			s.handleJobExeRes(jobExeRes)
		}
		//调度一次任务
		scheduleAfter = s.trySchedule()
		//重置一次调度器
		timer.Reset(scheduleAfter)
	}
}

// 处理任务事件
func (s *Scheduler) HandleJobEvent(event *JobEvent) {
	switch event.eventType {
	case common.JOB_EVENT_SAVE: //任务保存事件
		plan, err := s.buildSchedulePlan(event.job)
		if err != nil {
			return
		}
		Schedule.JobPlanMap[event.job.Name] = plan
	case common.JOB_EVENT_DELETE: // 任务删除事件
		if _, ok := Schedule.JobPlanMap[event.job.Name]; ok {
			delete(Schedule.JobPlanMap, event.job.Name)
		}
	case common.JOB_EVENT_KILL: //任务强杀事件
		//取消command执行  首先判断任务是否正在执行
		if exe, ok := Schedule.JobExecutingMap[event.job.Name]; ok {
			//触发command杀死shell子进程 任务退出
			fmt.Println("kill job : ", exe.Job.Name)
			exe.CancelFunc()
		}

	}
}

// 计算任务调度状态
func (s *Scheduler) trySchedule() time.Duration {
	now := time.Now()
	var nearTime *time.Time

	// 1. 遍历所有任务
	// 2. 过期的任务立即执行
	// 3. 统计最近的即将过期的任务的时间

	// 没有任务
	if len(s.JobPlanMap) == 0 {
		return 1 * time.Second
	}

	for _, plan := range s.JobPlanMap {
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			s.tryStartJob(plan)

			// 计算下次执行时间
			plan.NextTime = plan.Expr.Next(now)
		}

		// 统计最近要过期的任务时间
		if nearTime == nil || plan.NextTime.Before(*nearTime) {
			nearTime = &plan.NextTime
		}
	}

	// 下次调度时间间隔 = 最近要调度的时间 - 当前时间
	return (*nearTime).Sub(now)
}

//尝试执行任务
func (s *Scheduler) tryStartJob(plan *JobSchedulePlan) {
	//如果任务正在执行, 跳过这次调度
	if _, ok := s.JobExecutingMap[plan.Job.Name]; ok {
		fmt.Println("任务正在执行, 跳过本次执 : ", plan.Job.Name)
		return
	}

	//构建任务执行状态信息
	executeInfo := s.buildJobExecuteInfo(plan)
	//保存任务执行状态
	s.JobExecutingMap[executeInfo.Job.Name] = executeInfo

	//执行任务
	fmt.Println("执行任务 : ", executeInfo.Job.Name, executeInfo.PlanTime, executeInfo.RealTime)
	Exe.executeJob(executeInfo)
}

// 推送任务变化事件
func (s *Scheduler) PushJobEvent(event *JobEvent) {
	s.JobEventChan <- event
}

// 构建任务执行计划
func (s *Scheduler) buildSchedulePlan(job *Job) (*JobSchedulePlan, error) {
	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}

	return &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}, nil
}

//构建任务执行状态信息
func (s *Scheduler) buildJobExecuteInfo(plan *JobSchedulePlan) *JobExecuteInfo {
	jobExecuteInfo := &JobExecuteInfo{
		Job:      plan.Job,
		PlanTime: plan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.Ctx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())

	return jobExecuteInfo
}

//回传任务执行结果
func (s *Scheduler) pushJobExeRes(res *JobExecuteResult) {
	s.JobExeResChan <- res
}

//处理任务结果
func (s *Scheduler) handleJobExeRes(res *JobExecuteResult) {
	//从JobExecuteInfo中删除改条任务执行状态
	delete(s.JobExecutingMap, res.exeInfo.Job.Name)

	// 生成执行日志
	if res.err != common.ERR_LOCK_ALADY_REQUIRED {
		log := &JobLog{
			JobName:      res.exeInfo.Job.Name,
			Command:      res.exeInfo.Job.Command,
			OutPut:       string(res.output),
			PlanTime:     res.exeInfo.PlanTime.UnixNano() / 1000000,
			ScheduleTime: res.exeInfo.RealTime.UnixNano() / 1000000,
			StartTime:    res.startTime.UnixNano() / 1000000,
			EndTime:      res.endTime.UnixNano() / 1000000,
		}

		if res.err != nil {
			log.Error = res.err.Error()
		}
		// 将日志存储到MongoDB
		Sink.Append(log)
	}
}
