package uic

import (
	"github.com/astaxie/beego/orm"
	"github.com/open-falcon/fe/http/base"
	. "github.com/open-falcon/fe/model/uic"
	"github.com/open-falcon/fe/utils"
	"regexp"
	"strings"
)

type TeamController struct {
	base.BaseController
}

func (this *TeamController) Teams() {
	query := strings.TrimSpace(this.GetString("query", ""))
	if utils.HasDangerousCharacters(query) {
		this.ServeErrJson("query is invalid")
		return
	}

	per := this.MustGetInt("per", 10)
	me := this.Ctx.Input.GetData("CurrentUser").(*User)

	var teams orm.QuerySeter
	if me.Role == ROOT_ADMIN_ROLE {
		teams = QueryAllTeams(query)
	} else {
		var err error
		teams, err = QueryMineTeams(query, me.Id)
		if err != nil {
			this.ServeErrJson("occur error " + err.Error())
			return
		}
	}

	total, err := teams.Count()
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	pager := this.SetPaginator(per, total)
	teams = teams.Limit(per, pager.Offset())

	var ts []Team
	_, err = teams.All(&ts)
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	nteams := make([]map[string]interface{}, 0)
	for _, v := range ts {
		cu := ReadUserById(v.Creator)
		t := make(map[string]interface{})
		t["Id"] = v.Id
		t["Name"] = v.Name
		t["Resume"] = v.Resume
		t["CreatorCnname"] = cu.Cnname
		t["CreatorName"] = cu.Name
		t["IsAdmin"] = (v.IsAdmin(me.Id) || me.Role == ROOT_ADMIN_ROLE)
		nteams = append(nteams, t)
	}

	this.Data["Teams"] = nteams
	this.Data["Query"] = query
	this.Data["Me"] = me
	this.Data["IamRoot"] = me.Role == ROOT_ADMIN_ROLE
	this.TplName = "team/list.html"
}

func (this *TeamController) CreateTeamGet() {
	this.TplName = "team/create.html"
}

func (this *TeamController) CreateTeamPost() {
	name := strings.TrimSpace(this.GetString("name", ""))
	if name == "" {
		this.ServeErrJson("name is blank")
		return
	}

	if utils.HasDangerousCharacters(name) {
		this.ServeErrJson("name is invalid")
		return
	}

	resume := strings.TrimSpace(this.GetString("resume", ""))
	if utils.HasDangerousCharacters(resume) {
		this.ServeErrJson("resume is invalid")
		return
	}

	email := strings.TrimSpace(this.GetString("email", ""))
	if utils.HasDangerousCharacters(email) {
		this.ServeErrJson("email is invalid")
		return
	}
	if email != "" {
		mailArr := strings.Split(email, ",")
		for _, mailStr := range mailArr {
			if isOk, _ := regexp.MatchString("^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$", mailStr); !isOk {
				this.ServeErrJson("Email is invalid!")
				return
			}
		}
	}

	t := ReadTeamByName(name)
	if t != nil {
		this.ServeErrJson("name is already existent")
		return
	}

	sk := utils.RandStr(32)

	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	lastId, err := SaveTeamAttrs(name, resume, me.Id, email, sk)
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	if !me.IsRoot() {
		this.ServeErrJson("you are not root!")
		return
	}

	uids := strings.TrimSpace(this.GetString("users", ""))
	if utils.HasDangerousCharacters(uids) {
		this.ServeErrJson("uids is invalid")
		return
	}

	adminUids := strings.TrimSpace(this.GetString("admins", ""))
	if utils.HasDangerousCharacters(uids) {
		this.ServeErrJson("uids is invalid")
		return
	}

	err = PutUsersInTeam(lastId, uids)
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
	}

	uaerr := PutAdminInTeam(lastId, adminUids)
	this.AutoServeError(uaerr)
}

func (this *TeamController) Users() {
	teamName := strings.TrimSpace(this.GetString("name", ""))
	if teamName == "" {
		this.ServeErrJson("name is blank")
		return
	}

	this.Data["json"] = map[string]interface{}{
		"users": MembersByTeamName(teamName),
		"msg":   "",
	}
	this.ServeJSON()
}

func (this *TeamController) Admins() {
	teamName := strings.TrimSpace(this.GetString("name", ""))
	if teamName == "" {
		this.ServeErrJson("name is blank")
		return
	}

	this.Data["json"] = map[string]interface{}{
		"users": AdminsByTeamName(teamName),
		"msg":   "",
	}
	this.ServeJSON()
}

func (this *TeamController) DeleteTeam() {
	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	targetTeam := this.Ctx.Input.GetData("TargetTeam").(*Team)

	uid := me.Id
	if !targetTeam.IsAdmin(uid) && me.Role != ROOT_ADMIN_ROLE && targetTeam.Creator != me.Id {
		this.ServeErrJson("you are not admin")
		return
	}

	if !me.CanWrite(targetTeam) {
		this.ServeErrJson("no privilege")
		return
	}

	err := targetTeam.Remove()
	if err != nil {
		this.ServeErrJson(err.Error())
		return
	}

	this.ServeOKJson()
}

func (this *TeamController) EditGet() {
	targetTeam := this.Ctx.Input.GetData("TargetTeam").(*Team)
	user := ReadUserById(targetTeam.Creator)
	if user != nil {
		this.Data["TeamCreator"] = user.Name
	} else {
		this.Data["TeamCreator"] = "<Null>"
	}
	this.Data["TargetTeam"] = targetTeam

	loginUser := this.Ctx.Input.GetData("CurrentUser").(*User)
	uid := loginUser.Id
	this.Data["IsAdmin"] = (targetTeam.IsAdmin(uid) || loginUser.Role == ROOT_ADMIN_ROLE || targetTeam.Creator == loginUser.Id)
	this.TplName = "team/edit.html"
}

func (this *TeamController) EditPost() {
	targetTeam := this.Ctx.Input.GetData("TargetTeam").(*Team)
	resume := this.MustGetString("resume", "")
	userIdstr := this.MustGetString("users", "")
	teamEmail := this.MustGetString("teamemail", "")
	adminIdstr := this.MustGetString("admins", "")

	if utils.HasDangerousCharacters(resume) || utils.HasDangerousCharacters(teamEmail) || utils.HasDangerousCharacters(userIdstr) || utils.HasDangerousCharacters(userIdstr) {
		this.ServeErrJson("parameter resume or email or users or admins is invalid")
		return
	}

	if teamEmail != "" {
		mailArr := strings.Split(teamEmail, ",")
		for _, mailStr := range mailArr {
			if isOk, _ := regexp.MatchString("^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$", mailStr); !isOk {
				this.ServeErrJson("Email is invalid!")
				return
			}
		}
	}

	loginUser := this.Ctx.Input.GetData("CurrentUser").(*User)
	uid := loginUser.Id
	if !targetTeam.IsAdmin(uid) && loginUser.Role != ROOT_ADMIN_ROLE && targetTeam.Creator != loginUser.Id {
		this.ServeErrJson("you are not admin")
		return
	}

	if targetTeam.Resume != resume || targetTeam.Email != teamEmail {
		targetTeam.Resume = resume
		targetTeam.Email = teamEmail
		ClearTeamCacheById(targetTeam.Id)
		targetTeam.Update()
	}

	uuerr := targetTeam.UpdateUsers(userIdstr)
	if uuerr != nil {
		this.AutoServeError(uuerr)
	}

	uaerr := targetTeam.UpdateAdmins(adminIdstr)
	this.AutoServeError(uaerr)

}

// for portal api: query team
func (this *TeamController) Query() {
	query := this.MustGetString("query", "")
	limit := this.MustGetInt("limit", 10)

	qs := QueryAllTeams(query)
	var ts []Team
	qs.Limit(limit).All(&ts)
	this.Data["json"] = map[string]interface{}{
		"msg":   "",
		"teams": ts,
	}
	this.ServeJSON()
}

func (this *TeamController) All() {
	this.Redirect("/me/teams", 301)
}

func (this *TeamController) Checksk() {
	team := strings.TrimSpace(this.GetString("team", ""))
	sk := strings.TrimSpace(this.GetString("secretkey", ""))
	if team == "" || sk == "" {
		this.Ctx.Output.Body([]byte("-1"))
		return
	}

	if utils.HasDangerousCharacters(team) || utils.HasDangerousCharacters(sk) {
		this.Ctx.Output.Body([]byte("-2"))
		return
	}

	t := ReadTeamByName(team)
	if t == nil {
		this.Ctx.Output.Body([]byte("-1"))
		return
	}

	if t.Secretkey == sk {
		this.Ctx.Output.Body([]byte("1"))
	} else {
		this.Ctx.Output.Body([]byte("0"))
	}
}
