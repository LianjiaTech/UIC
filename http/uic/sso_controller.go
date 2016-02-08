package uic

import (
	"github.com/open-falcon/fe/http/base"
	"github.com/open-falcon/fe/model/uic"
	"github.com/open-falcon/fe/utils"
)

type SsoController struct {
	base.BaseController
}

func (this *SsoController) Sig() {
	this.Ctx.Output.Body([]byte(utils.GenerateUUID()))
}

func (this *SsoController) User() {
	sig := this.Ctx.Input.Param(":sig")
	if sig == "" {
		this.NotFound("sig is blank")
		return
	}

	s := uic.ReadSessionBySig(sig)
	if s == nil {
		this.NotFound("no such sig")
		return
	}

	u := uic.ReadUserById(s.Uid)
	if u == nil {
		this.NotFound("no such user")
		return
	}

	this.Data["json"] = map[string]interface{}{
		"user": u,
	}
	this.ServeJSON()
}

func (this *SsoController) Logout() {
	sig := this.Ctx.Input.Param(":sig")
	if sig == "" {
		this.ServeErrJson("sig is blank")
		return
	}

	s := uic.ReadSessionBySig(sig)
	if s != nil {
		uic.RemoveSessionByUid(s.Uid)
	}

	this.ServeOKJson()
}
