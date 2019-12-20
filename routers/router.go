package routers

import (
	"etcd-learn/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/*", &controllers.MainController{})

	beego.Router("/job/save", &controllers.MasterController{}, "post:Save")
	beego.Router("/job/delete", &controllers.MasterController{}, "post:Delete")
	beego.Router("/job/jobList", &controllers.MasterController{}, "get:JobList")
	beego.Router("/job/killJob", &controllers.MasterController{}, "post:KillJob")
	beego.Router("/job/log", &controllers.MasterController{}, "post:JobLog")
	beego.Router("/worker/list", &controllers.MasterController{}, "get:WorkList")
}
