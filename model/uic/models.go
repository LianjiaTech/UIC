package uic

import (
	"time"
)

const (
    ROOT_ADMIN_ROLE int = 2
    TEAM_ADMIN_ROLE int = 1
    NOMAIL_ROLE int = 0
)

type User struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name"`
	Cnname  string    `json:"cnname"`
	Passwd  string    `json:"-"`
	Email   string    `json:"email"`
	Phone   string    `json:"phone"`
	IM      string    `json:"im" orm:"column(im)"`
	QQ      string    `json:"qq" orm:"column(qq)"`
	Role    int       `json:"role"`
	Created time.Time `json:"-"`
}

type Team struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name"`
	Resume  string    `json:"resume"`
	Creator int64     `json:"creator"`
    Email   string    `json:"email"`
    Secretkey   string  `json:"secretkey"`
	Created time.Time `json:"-"`
}

type RelTeamUser struct {
	Id  int64
	Tid int64
	Uid int64
    Isadmin bool    `json:"is_admin"`
}

type Session struct {
	Id      int64
	Uid     int64
	Sig     string
	Expired int
}
