package http

import (
	"github.com/astaxie/beego"
	"github.com/open-falcon/fe/g"
	"github.com/open-falcon/fe/http/home"
	"github.com/open-falcon/fe/http/uic"
	uic_model "github.com/open-falcon/fe/model/uic"
)

func Start() {
	if !g.Config().Http.Enabled {
		return
	}

	addr := g.Config().Http.Listen
	if addr == "" {
		return
	}

	home.ConfigRoutes()
	uic.ConfigRoutes()

	beego.AddFuncMap("member", uic_model.MembersByTeamId)
	beego.Run(addr)
}
