package uic

import (
	"github.com/open-falcon/fe/g"
	"github.com/open-falcon/fe/http/base"
	. "github.com/open-falcon/fe/model/uic"
	"github.com/open-falcon/fe/utils"
	"github.com/toolkits/rsc/qr"
	"github.com/toolkits/str"
	"github.com/toolkits/web"
	"strings"
	"time"
)

type UserController struct {
	base.BaseController
}

func (this *UserController) CreateRoot() {
	password := strings.TrimSpace(this.GetString("password", ""))
	if password == "" {
		this.Ctx.Output.Body([]byte("password is blank"))
		return
	}

	userPtr := &User{
		Name:    "root",
		Passwd:  str.Md5Encode(g.Config().Salt + password),
		Role:    2,
		Created: time.Now(),
	}

	_, err := userPtr.Save()
	if err != nil {
		this.Ctx.Output.Body([]byte(err.Error()))
	} else {
		this.Ctx.Output.Body([]byte("success"))
	}
}

// 登录成功之后跳转的欢迎页面
func (this *UserController) Info() {
	this.Data["CurrentUser"] = this.Ctx.Input.GetData("CurrentUser").(*User)
	this.TplName = "user/info.html"
}

// 展示当前用户的联系信息
func (this *UserController) ProfileGet() {
	this.Data["CurrentUser"] = this.Ctx.Input.GetData("CurrentUser").(*User)
	this.TplName = "user/profile.html"
}

// 更新个人信息
func (this *UserController) ProfilePost() {
	im := strings.TrimSpace(this.GetString("im", ""))
	qq := strings.TrimSpace(this.GetString("qq", ""))
    cnname := strings.TrimSpace(this.GetString("cnname",""))
    phone := strings.TrimSpace(this.GetString("phone",""))

	if utils.HasDangerousCharacters(im) {
		this.ServeErrJson("im is invalid")
		return
	}

	if utils.HasDangerousCharacters(qq) {
		this.ServeErrJson("qq is invalid")
		return
	}

	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	me.IM = im
	me.QQ = qq
    me.Cnname = cnname
    me.Phone = phone

	me.Update()
	this.ServeOKJson()
}

// 为某个用户设置角色，要求当前用户得是root
func (this *UserController) Role() {
	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	if me.Role != ROOT_ADMIN_ROLE {
		this.ServeErrJson("no privilege")
		return
	}

	targetUser := this.Ctx.Input.GetData("TargetUser").(*User)

	if targetUser.Name == "root" {
		this.ServeErrJson("no privilege")
		return
	}

	newRole, err := this.GetInt("role", -1)
	if err != nil || newRole == -1 {
		this.ServeErrJson("parameter role is necessary")
		return
	}

	targetUser.Role = newRole
	_, err = targetUser.Update()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *UserController) ChangePassword() {
	oldPassword := strings.TrimSpace(this.GetString("old_password", ""))
	newPassword := strings.TrimSpace(this.GetString("new_password", ""))
	repeatPassword := strings.TrimSpace(this.GetString("repeat_password", ""))

	if newPassword != repeatPassword {
		this.ServeErrJson("password not equal the repeart one")
		return
	}

	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	if me.Passwd != str.Md5Encode(g.Config().Salt+oldPassword) {
		this.ServeErrJson("old password error")
		return
	}

	newPass := str.Md5Encode(g.Config().Salt + newPassword)
	if me.Passwd == newPass {
		this.ServeOKJson()
		return
	}

	me.Passwd = newPass
	_, err := me.Update()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	RemoveSessionByUid(me.Id)
	this.ServeOKJson()
}

func (this *UserController) Users() {
	query := strings.TrimSpace(this.GetString("query", ""))
	if utils.HasDangerousCharacters(query) {
		this.ServeErrJson("query is invalid")
		return
	}

	var us []User
	var total int64
	per := this.MustGetInt("per", 20)
	var pager *web.Paginator

	if !g.Config().Ldap.Enabled {
		users := QueryUsers(query)
		total, err := users.Count()
		if err != nil {
			this.ServeErrJson("occur error " + err.Error())
			return
		}

		pager = this.SetPaginator(per, total)
		users = users.Limit(per, pager.Offset())

		_, err = users.All(&us)
		if err != nil {
			this.ServeErrJson("occur error " + err.Error())
			return
		}
	} else {
		user_attributes, err := utils.Ldapsearch(g.Config().Ldap.Addr,
			g.Config().Ldap.BaseDN,
			g.Config().Ldap.BindDN,
			g.Config().Ldap.BindPasswd,
			g.Config().Ldap.UserField,
			query,
			g.Config().Ldap.Attributes)
		userSn := ""
		userMail := ""
		userTel := ""
		if err == nil {
			userSn = user_attributes["sn"]
			userMail = user_attributes["mail"]
			userTel = user_attributes["telephoneNumber"]

			u := User{
				Name:   query,
				Passwd: "",
				Cnname: userSn,
				Phone:  userTel,
				Email:  userMail,
			}
			total = 1

			//查询此用户的role
			obj := ReadUserByName(query)
			if obj == nil {
				if userSn != "" {
					// 说明用户不存在
					obj = &User{
						Name:    query,
						Passwd:  "",
						Cnname:  userSn,
						Phone:   userTel,
						Email:   userMail,
						Created: time.Now(),
					}
					_, err = obj.Save()
					if err != nil {
						this.ServeErrJson("insert user fail " + err.Error())
						return
					}
				}
			} else {
				u.Role = obj.Role
				u.QQ = obj.QQ
				u.IM = obj.IM
			}
			us = append(us, u)
		}
		pager = this.SetPaginator(per, total)
	}

	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	this.Data["Users"] = us
	this.Data["Query"] = query
	this.Data["Me"] = me
	this.Data["IamRoot"] = me.Role == ROOT_ADMIN_ROLE
	this.TplName = "user/list.html"
}

func (this *UserController) CreateUserGet() {
	this.TplName = "user/create.html"
}

func (this *UserController) CreateUserPost() {
	name := strings.TrimSpace(this.GetString("name", ""))
	password := strings.TrimSpace(this.GetString("password", ""))
	cnname := strings.TrimSpace(this.GetString("cnname", ""))
	email := strings.TrimSpace(this.GetString("email", ""))
	phone := strings.TrimSpace(this.GetString("phone", ""))
	im := strings.TrimSpace(this.GetString("im", ""))
	qq := strings.TrimSpace(this.GetString("qq", ""))

	if !utils.IsUsernameValid(name) {
		this.ServeErrJson("name pattern is invalid")
		return
	}

	if ReadUserIdByName(name) > 0 {
		this.ServeErrJson("name is already existent")
		return
	}

	if password == "" {
		this.ServeErrJson("password is blank")
		return
	}

	if utils.HasDangerousCharacters(cnname) {
		this.ServeErrJson("cnname is invalid")
		return
	}

	if utils.HasDangerousCharacters(email) {
		this.ServeErrJson("email is invalid")
		return
	}

	if utils.HasDangerousCharacters(phone) {
		this.ServeErrJson("phone is invalid")
		return
	}

	if utils.HasDangerousCharacters(im) {
		this.ServeErrJson("im is invalid")
		return
	}

	if utils.HasDangerousCharacters(qq) {
		this.ServeErrJson("qq is invalid")
		return
	}

	lastId, err := InsertRegisterUser(name, str.Md5Encode(g.Config().Salt+password))
	if err != nil {
		this.ServeErrJson("insert user fail " + err.Error())
		return
	}

	targetUser := ReadUserById(lastId)
	targetUser.Cnname = cnname
	targetUser.Email = email
	targetUser.Phone = phone
	targetUser.IM = im
	targetUser.QQ = qq

	if _, err := targetUser.Update(); err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *UserController) DeleteUser() {
	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	if me.Role <= 0 {
		this.ServeErrJson("no privilege")
		return
	}

	userPtr := this.Ctx.Input.GetData("TargetUser").(*User)

	_, err := userPtr.Remove()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *UserController) EditGet() {
	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	if me.Role <= 0 {
		this.ServeErrJson("no privilege")
		return
	}

	this.Data["User"] = this.Ctx.Input.GetData("TargetUser").(*User)
	this.TplName = "user/edit.html"
}

func (this *UserController) EditPost() {
	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	if me.Role != ROOT_ADMIN_ROLE {
		this.ServeErrJson("no privilege")
		return
	}
	cnname := strings.TrimSpace(this.GetString("cnname", ""))
	email := strings.TrimSpace(this.GetString("email", ""))
	phone := strings.TrimSpace(this.GetString("phone", ""))
	im := strings.TrimSpace(this.GetString("im", ""))
	qq := strings.TrimSpace(this.GetString("qq", ""))

	if utils.HasDangerousCharacters(cnname) {
		this.ServeErrJson("cnname is invalid")
		return
	}

	if utils.HasDangerousCharacters(email) {
		this.ServeErrJson("email is invalid")
		return
	}

	if utils.HasDangerousCharacters(phone) {
		this.ServeErrJson("phone is invalid")
		return
	}

	if utils.HasDangerousCharacters(im) {
		this.ServeErrJson("im is invalid")
		return
	}

	if utils.HasDangerousCharacters(qq) {
		this.ServeErrJson("qq is invalid")
		return
	}

	targetUser := this.Ctx.Input.GetData("TargetUser").(*User)
	if targetUser.Name == "root" {
		this.ServeErrJson("no privilege")
		return
	}

	targetUser.Cnname = cnname
	targetUser.Email = email
	targetUser.Phone = phone
	targetUser.IM = im
	targetUser.QQ = qq

	_, err := targetUser.Update()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *UserController) ResetPassword() {
	password := this.GetString("password", "")
	if password == "" {
		this.ServeErrJson("password is blank")
		return
	}

	targetUser := this.Ctx.Input.GetData("TargetUser").(*User)
	if targetUser.Name == "root" {
		this.ServeErrJson("no privilege")
		return
	}

	targetUser.Passwd = str.Md5Encode(g.Config().Salt + password)
	_, err := targetUser.Update()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *UserController) Query() {
	query := strings.TrimSpace(this.GetString("query", ""))
	query = strings.ToLower(query)
	limit := this.MustGetInt("limit", 10)

	if utils.HasDangerousCharacters(query) {
		this.ServeErrJson("query is invalid")
		return
	}

	var users []User
	QueryUsers(query).Limit(limit).All(&users, "Id", "Name", "Cnname", "Email", "Phone")

	isInLdap := false
	for _, v := range users {
		if strings.ToLower(v.Name) == query {
			isInLdap = true
		}
	}

	if isInLdap == false {
		user_attributes, err := utils.Ldapsearch(g.Config().Ldap.Addr,
			g.Config().Ldap.BaseDN,
			g.Config().Ldap.BindDN,
			g.Config().Ldap.BindPasswd,
			g.Config().Ldap.UserField,
			query,
			g.Config().Ldap.Attributes)
		userSn := ""
		userMail := ""
		userTel := ""
		if err == nil && len(user_attributes) > 0 {
			userSn = user_attributes["sn"]
			userMail = user_attributes["mail"]
			userTel = user_attributes["telephoneNumber"]

			u := ReadUserByName(query)
			if u == nil {
				// 说明用户不存在
				u = &User{
					Name:    query,
					Passwd:  "",
					Cnname:  userSn,
					Phone:   userTel,
					Email:   userMail,
					Created: time.Now(),
				}
				_, err = u.Save()
				if err != nil {
					this.ServeErrJson("insert user fail " + err.Error())
					return
				}
			}

			users = append(users, *u)
		}
	}

	this.Data["json"] = map[string]interface{}{"users": users}
	this.ServeJSON()
}

func (this *UserController) In() {
	name := this.MustGetString("name", "")
	teamNames := this.MustGetString("teams", "")

	if name == "" || teamNames == "" {
		this.Ctx.Output.Body([]byte("0"))
		return
	}

	teamNames = strings.Replace(teamNames, ";", ",", -1)
	teamArr := strings.Split(teamNames, ",")
	for _, teamName := range teamArr {
		t := ReadTeamByName(teamName)
		if t == nil {
			continue
		}

		members := MembersByTeamName(teamName)
		for _, u := range members {
			if u.Name == name {
				this.Ctx.Output.Body([]byte("1"))
				return
			}
		}
	}

	this.Ctx.Output.Body([]byte("0"))
}

func (this *UserController) About() {
	name := this.Ctx.Input.Param(":name")
	var u *User
	if !g.Config().Ldap.Enabled {
		u = ReadUserByName(name)
	} else {
		user_attributes, err := utils.Ldapsearch(g.Config().Ldap.Addr,
			g.Config().Ldap.BaseDN,
			g.Config().Ldap.BindDN,
			g.Config().Ldap.BindPasswd,
			g.Config().Ldap.UserField,
			name,
			g.Config().Ldap.Attributes)
		userSn := ""
		userMail := ""
		userTel := ""
		if err == nil {
			userSn = user_attributes["sn"]
			userMail = user_attributes["mail"]
			userTel = user_attributes["telephoneNumber"]

			u = &User{
				Name:    name,
				Passwd:  "",
				Cnname:  userSn,
				Phone:   userTel,
				Email:   userMail,
				Created: time.Now(),
			}

			udb := ReadUserByName(name)
			if udb != nil {
				u.QQ = udb.QQ
				u.IM = udb.IM
			}
		}
	}

	if u == nil {
		this.NotFound("no such user")
		return
	}

	this.Data["User"] = u
	this.TplName = "user/about.html"
}

func (this *UserController) QrCode() {
	name := this.Ctx.Input.Param(":name")
	u := ReadUserByName(name)
	if u == nil {
		this.NotFound("no such user")
		return
	}

	c, err := qr.Encode("BEGIN:VCARD\nVERSION:3.0\nFN:"+u.Cnname+"\nTEL;WORK;VOICE:"+u.Phone+"\nEMAIL;PREF;INTERNET:"+u.Email+"\nORG:"+g.Config().Company+"\nEND:VCARD", qr.L)
	if err != nil {
		this.NotFound("no such user")
		return
	}

	this.Ctx.Output.ContentType("image")
	this.Ctx.Output.Body(c.PNG())
}

func (this *UserController) Teams() {
	userName := strings.TrimSpace(this.GetString("name", ""))
	if userName == "" {
		this.ServeErrJson("name is blank")
		return
	}

	this.Data["json"] = map[string]interface{}{
		"teams": GetTeamsByUserName(userName),
		"msg":   "",
	}
	this.ServeJSON()

}

func (this *UserController) TeamsAdmin() {
	userName := strings.TrimSpace(this.GetString("name", ""))
	teamNames := strings.TrimSpace(this.GetString("teams", ""))
	if userName == "" || teamNames == "" {
		this.ServeErrJson("name is blank")
	}

	userObj := ReadUserByName(userName)
	if userObj == nil {
		this.ServeErrJson("The User is not exists!")
	}

	res := make(map[string]interface{})
	teamArr := strings.Split(teamNames, ",")
	for _, teamName := range teamArr {
		t := ReadTeamByName(teamName)
		if t == nil {
			continue
		}

		res[teamName] = (t.IsAdmin(userObj.Id) || userObj.Role == ROOT_ADMIN_ROLE || t.Creator == userObj.Id)
	}

	this.Data["json"] = map[string]interface{}{
		"admin": res,
		"msg":   "",
	}
	this.ServeJSON()

}
