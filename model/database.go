package model

import (
	"github.com/astaxie/beego/orm"
	"github.com/open-falcon/fe/g"
	. "github.com/open-falcon/fe/model/uic"
	_ "github.com/go-sql-driver/mysql"
)

func InitDatabase() {
	// set default database
	config := g.Config()
	orm.RegisterDataBase("default", "mysql", config.Uic.Addr, config.Uic.Idle, config.Uic.Max)

	// register model
	orm.RegisterModel(new(User), new(Team), new(Session), new(RelTeamUser))

	if config.Log == "debug" {
		orm.Debug = true
	}
}
