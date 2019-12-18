package controllers

import (
	"encoding/json"
	. "etcd-learn/crontab/master"
	"fmt"
	"github.com/astaxie/beego"
)

type MasterController struct {
	beego.Controller
}

/*
保存新增的job任务

{
"name" : "job1",
"command" : "echo hello",
"cronExpr" : ""
}
 */
func (c *MasterController) Save() {
	var job Job

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &job); err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}

	fmt.Println(job)

	if oldJob, err := job.SaveJob(); err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	} else {
		c.Data["json"] = Response{Code: 200, Message: "success", Data: oldJob}
		c.ServeJSON()
	}
}

/*
删除任务

{
"name" : "job1"
}
 */

func (c *MasterController) Delete() {
	var job Job

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &job); err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}

	old, err := job.DeleteJob()
	if err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = Response{Code: 200, Message: "success", Data: old}
	c.ServeJSON()
}

//返回所有的任务列表
func (c *MasterController) JobList() {
	var job Job
	jobs, err := job.JobList()
	if err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = Response{Code: 200, Message: "success", Data: jobs}
	c.ServeJSON()
}

/*
kill任务

{
"name" : "job1"
}
 */
func (c *MasterController) KillJob() {
	var job Job

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &job); err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}

	if err := job.KillJob(); err != nil {
		c.Data["json"] = Response{Code: 500, Message: err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = Response{Code: 200, Message: "success"}
	c.ServeJSON()
}