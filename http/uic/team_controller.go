package uic

import (
	"github.com/astaxie/beego/orm"
	"github.com/open-falcon/fe/http/base"
	. "github.com/open-falcon/fe/model/uic"
	"github.com/open-falcon/fe/utils"
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

    nteams := make([]map[string]interface{},0)
    for _,v := range ts {
        cu := ReadUserById(v.Creator)
        t := make(map[string]interface{})
        t["Id"] = v.Id
        t["Name"] = v.Name
        t["Resume"] = v.Resume
        t["CreatorCnname"] = cu.Cnname
        t["CreatorName"] = cu.Name
        nteams = append(nteams,t)
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

	t := ReadTeamByName(name)
	if t != nil {
		this.ServeErrJson("name is already existent")
		return
	}

	me := this.Ctx.Input.GetData("CurrentUser").(*User)
	lastId, err := SaveTeamAttrs(name, resume, me.Id)
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
		return
	}

	uids := strings.TrimSpace(this.GetString("users", ""))
	if utils.HasDangerousCharacters(uids) {
		this.ServeErrJson("uids is invalid")
		return
	}

	err = PutUsersInTeam(lastId, uids)
	if err != nil {
		this.ServeErrJson("occur error " + err.Error())
	} else {
		this.ServeOKJson()
	}
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
    user := ReadUserById( targetTeam.Creator )
    if user != nil {
        this.Data["TeamCreator"] = user.Name   
    } else {
        this.Data["TeamCreator"] = "<Null>" 
    }
	this.Data["TargetTeam"] = targetTeam
	this.TplName = "team/edit.html"
}

func (this *TeamController) EditPost() {
	targetTeam := this.Ctx.Input.GetData("TargetTeam").(*Team)
	resume := this.MustGetString("resume", "")
	userIdstr := this.MustGetString("users", "")
	adminIdstr := this.MustGetString("admins", "")

	if utils.HasDangerousCharacters(resume) || utils.HasDangerousCharacters(userIdstr) || utils.HasDangerousCharacters(userIdstr) {
		this.ServeErrJson("parameter resume or users or admins is invalid")
		return
	}

	if targetTeam.Resume != resume {
		targetTeam.Resume = resume
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
