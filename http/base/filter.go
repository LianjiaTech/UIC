package base

import (
	"github.com/astaxie/beego/context"
	"github.com/open-falcon/fe/g"
	"github.com/open-falcon/fe/model/uic"
	"github.com/open-falcon/fe/utils"
	"strconv"
	"time"
)

var FilterLoginUser = func(ctx *context.Context) {
	cookieSig := ctx.GetCookie("sig")
	if cookieSig == "" {
		ctx.Redirect(302, "/auth/login?callback="+ctx.Request.URL.String())
		return
	}

	sessionObj := uic.ReadSessionBySig(cookieSig)
	if sessionObj == nil || int64(sessionObj.Expired) < time.Now().Unix() {
		ctx.Redirect(302, "/auth/login?callback="+ctx.Request.URL.String())
		return
	}

	u := uic.ReadUserById(sessionObj.Uid)
	if u == nil {
		ctx.Redirect(302, "/auth/login?callback="+ctx.Request.URL.String())
		return
	}

	ctx.Input.SetData("CurrentUser", u)
}

var FilterTargetUser = func(ctx *context.Context) {
	userName := ctx.Input.Query("name")
	if userName == "" {
		ctx.ResponseWriter.WriteHeader(403)
		ctx.ResponseWriter.Write([]byte("Name is necessary"))
		return
	}

	u := uic.ReadUserByName(userName)
	if u == nil {
		user_attributes, err := utils.Ldapsearch(g.Config().Ldap.Addr,
			g.Config().Ldap.BaseDN,
			g.Config().Ldap.BindDN,
			g.Config().Ldap.BindPasswd,
			g.Config().Ldap.UserField,
			userName,
			g.Config().Ldap.Attributes)
		userSn := ""
		userMail := ""
		userTel := ""

		if err == nil {
			userSn = user_attributes["sn"]
			userMail = user_attributes["mail"]
			userTel = user_attributes["mobile"]
		}

		u = &uic.User{
			Name:   userName,
			Passwd: "",
			Cnname: userSn,
			Phone:  userTel,
			Email:  userMail,
            }
        u.Role = uic.NOMAIL_ROLE
		_, err = u.Save()
		if err != nil {
			ctx.ResponseWriter.WriteHeader(403)
			ctx.ResponseWriter.Write([]byte("insert user failed"))
			return
		}
	}

	ctx.Input.SetData("TargetUser", u)
}

var FilterTargetTeam = func(ctx *context.Context) {
	tid := ctx.Input.Query("id")
	if tid == "" {
		ctx.ResponseWriter.WriteHeader(403)
		ctx.ResponseWriter.Write([]byte("id is necessary"))
		return
	}

	id, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		ctx.ResponseWriter.WriteHeader(403)
		ctx.ResponseWriter.Write([]byte("id is invalid"))
		return
	}

	t := uic.ReadTeamById(id)
	if t == nil {
		ctx.ResponseWriter.WriteHeader(403)
		ctx.ResponseWriter.Write([]byte("no such team"))
		return
	}

	ctx.Input.SetData("TargetTeam", t)
}
