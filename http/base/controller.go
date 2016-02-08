package base

import (
	"github.com/astaxie/beego"
	"github.com/toolkits/web"
	"strings"
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) ServeErrJson(msg string) {
	this.Data["json"] = map[string]interface{}{
		"msg": msg,
	}
	this.ServeJSON()
}

func (this *BaseController) SetPaginator(per int, nums int64) *web.Paginator {
	p := web.NewPaginator(this.Ctx.Request, per, nums)
	this.Data["paginator"] = p
	return p
}

func (this *BaseController) ServeOKJson() {
	this.Data["json"] = map[string]interface{}{
		"msg": "",
	}
	this.ServeJSON()
}

func (this *BaseController) AutoServeError(err error) {
	if err != nil {
		this.ServeErrJson(err.Error())
	} else {
		this.ServeOKJson()
	}
}

func (this *BaseController) ServeDataJson(data interface{}) {
	this.Data["json"] = map[string]interface{}{
		"msg":  "",
		"data": data,
	}
	this.ServeJSON()
}

func (this *BaseController) NotFound(body string) {
	this.Ctx.ResponseWriter.WriteHeader(404)
	this.Ctx.ResponseWriter.Write([]byte(body))
}

func (this *BaseController) MustGetInt(key string, def int) int {
	val, err := this.GetInt(key, def)
	if err != nil {
		return def
	}

	return val
}

func (this *BaseController) MustGetInt64(key string, def int64) int64 {
	val, err := this.GetInt64(key, def)
	if err != nil {
		return def
	}

	return val
}

func (this *BaseController) MustGetString(key string, def string) string {
	return strings.TrimSpace(this.GetString(key, def))
}

func (this *BaseController) MustGetBool(key string, def bool) bool {
	raw := strings.TrimSpace(this.GetString(key, "0"))
	if raw == "true" || raw == "1" || raw == "on" || raw == "checked" || raw == "yes" {
		return true
	} else if raw == "false" || raw == "0" || raw == "off" || raw == "" || raw == "no" {
		return false
	} else {
		return def
	}
}
