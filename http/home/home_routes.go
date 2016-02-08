package home

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/open-falcon/fe/g"
)

func ConfigRoutes() {
	beego.Router("/", &HomeController{})

	beego.Get("/health", func(ctx *context.Context) {
		ctx.Output.Body([]byte("ok"))
	})

	beego.Get("/version", func(ctx *context.Context) {
		ctx.Output.Body([]byte(g.VERSION))
	})
}
