package uic

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/open-falcon/fe/http/base"
)

func ConfigRoutes() {

	beego.Router("/root", &UserController{}, "get:CreateRoot")

	beego.Router("/auth/login", &AuthController{}, "get:LoginGet;post:LoginPost")
	beego.Router("/auth/register", &AuthController{}, "get:RegisterGet;post:RegisterPost")

	beego.Router("/sso/sig", &SsoController{}, "get:Sig")
	beego.Router("/sso/user/:sig:string", &SsoController{}, "get:User")
	beego.Router("/sso/logout/:sig:string", &SsoController{}, "get:Logout")

	beego.Router("/user/query", &UserController{}, "get:Query")
	beego.Router("/user/teams", &UserController{}, "get:Teams")
	beego.Router("/user/teamadmin", &UserController{}, "get:TeamsAdmin")
	beego.Router("/user/in", &UserController{}, "get:In")
	beego.Router("/user/qrcode/:name:string", &UserController{}, "get:QrCode")
	beego.Router("/about/:name:string", &UserController{}, "get:About")

	beego.Router("/team/users", &TeamController{}, "get:Users")
	beego.Router("/team/admins", &TeamController{}, "get:Admins")
	beego.Router("/team/query", &TeamController{}, "get:Query")
	beego.Router("/team/checksk", &TeamController{}, "get:Checksk")
	beego.Router("/team/all", &TeamController{}, "get:All")

	loginRequired :=
		beego.NewNamespace("/me",
			beego.NSCond(func(ctx *context.Context) bool {
				return true
			}),
			beego.NSBefore(base.FilterLoginUser),
			beego.NSRouter("/logout", &AuthController{}, "*:Logout"),
			beego.NSRouter("/info", &UserController{}, "get:Info"),
			beego.NSRouter("/profile", &UserController{}, "get:ProfileGet;post:ProfilePost"),
			beego.NSRouter("/chpwd", &UserController{}, "*:ChangePassword"),
			beego.NSRouter("/users", &UserController{}, "get:Users"),
			beego.NSRouter("/user/c", &UserController{}, "get:CreateUserGet;post:CreateUserPost"),
			beego.NSRouter("/teams", &TeamController{}, "get:Teams"),
			beego.NSRouter("/team/c", &TeamController{}, "get:CreateTeamGet;post:CreateTeamPost"),
		)

	beego.AddNamespace(loginRequired)

	targetUserRequired :=
		beego.NewNamespace("/target-user",
			beego.NSCond(func(ctx *context.Context) bool {
				return true
			}),
			beego.NSBefore(base.FilterLoginUser, base.FilterTargetUser),
			beego.NSRouter("/delete", &UserController{}, "*:DeleteUser"),
			beego.NSRouter("/edit", &UserController{}, "get:EditGet;post:EditPost"),
			beego.NSRouter("/chpwd", &UserController{}, "post:ResetPassword"),
			beego.NSRouter("/role", &UserController{}, "*:Role"),
		)

	beego.AddNamespace(targetUserRequired)

	targetTeamRequired :=
		beego.NewNamespace("/target-team",
			beego.NSCond(func(ctx *context.Context) bool {
				return true
			}),
			beego.NSBefore(base.FilterLoginUser, base.FilterTargetTeam),
			beego.NSRouter("/delete", &TeamController{}, "*:DeleteTeam"),
			beego.NSRouter("/edit", &TeamController{}, "get:EditGet;post:EditPost"),
		)

	beego.AddNamespace(targetTeamRequired)

}
