package uic

import (
	"github.com/astaxie/beego/orm"
	"github.com/toolkits/cache"
	"github.com/toolkits/logger"
	"strconv"
	"strings"
	"time"
    "fmt"
)

func QueryMineTeams(query string, uid int64) (orm.QuerySeter, error) {
	qs := orm.NewOrm().QueryTable(new(Team))

	condMine := orm.NewCondition()
	condMine = condMine.Or("Creator", uid)

	tids, err := Tids(uid)
	if err != nil {
		return qs, err
	}

	if len(tids) > 0 {
		condMine = condMine.Or("Id__in", tids)
	}

	condResult := orm.NewCondition().AndCond(condMine)

	if query != "" {
		condQuery := orm.NewCondition()
		condQuery = condQuery.And("Name__icontains", query)
		condResult = condResult.AndCond(condQuery)
	}

	qs = qs.SetCond(condResult)
	return qs, nil
}

func QueryAllTeams(query string) orm.QuerySeter {
	qs := orm.NewOrm().QueryTable(new(Team))
	if query != "" {
		qs = qs.Filter("Name__icontains", query)
	}
	return qs
}

func Tids(uid int64) ([]int64, error) {
	type TidStruct struct {
		Tid int64
	}

	var tids []TidStruct
	_, err := orm.NewOrm().Raw("select tid from rel_team_user where uid = ?", uid).QueryRows(&tids)
	if err != nil {
		return []int64{}, err
	}

	size := len(tids)
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = tids[i].Tid
	}

	return arr, nil
}

func Uids(tid int64) ([]int64, error) {
	type UidStruct struct {
		Uid int64
	}

	var uids []UidStruct
	_, err := orm.NewOrm().Raw("select uid from rel_team_user where tid = ?", tid).QueryRows(&uids)
	if err != nil {
		return []int64{}, err
	}

	size := len(uids)
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = uids[i].Uid
	}

	return arr, nil
}

func AdminUids(tid int64) ([]int64, error) {
	type UidStruct struct {
		Uid int64
	}

	var uids []UidStruct
	_, err := orm.NewOrm().Raw("select uid from rel_team_user where is_admin=1 and tid = ?", tid).QueryRows(&uids)
	if err != nil {
		return []int64{}, err
	}

	size := len(uids)
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = uids[i].Uid
	}

	return arr, nil
}

func SelectTeamById(id int64) *Team {
	if id <= 0 {
		return nil
	}

	obj := Team{Id: id}
	err := orm.NewOrm().Read(&obj, "Id")
	if err != nil {
		if err != orm.ErrNoRows {
			logger.Errorln(err)
		}
		return nil
	}
	return &obj
}

func ReadTeamById(id int64) *Team {
	if id <= 0 {
		return nil
	}

	key := fmt.Sprintf("team:obj:%d", id)
	var obj Team
	if err := cache.Get(key, &obj); err != nil {
		objPtr := SelectTeamById(id)
		if objPtr != nil {
			go cache.Set(key, objPtr, time.Hour)
		}
		return objPtr
	}

	return &obj
}

func ClearTeamCacheById(id int64) {
    key := fmt.Sprintf("team:obj:%d", id )   
    go cache.Delete(key)
}

func SelectTeamIdByName(name string) int64 {
	if name == "" {
		return 0
	}

	type IdStruct struct {
		Id int64
	}

	var idObj IdStruct
	err := orm.NewOrm().Raw("select id from team where name = ?", name).QueryRow(&idObj)
	if err != nil {
		return 0
	}

	return idObj.Id
}

func ReadTeamIdByName(name string) int64 {
	if name == "" {
		return 0
	}

	key := fmt.Sprintf("team:id:%s", name)
	var id int64
	if err := cache.Get(key, &id); err != nil {
		id = SelectTeamIdByName(name)
		if id > 0 {
			go cache.Set(key, id, time.Hour)
		}
	}

	return id
}

func ReadTeamByName(name string) *Team {
	return ReadTeamById(ReadTeamIdByName(name))
}

func (this *Team) Save() (int64, error) {
	return orm.NewOrm().Insert(this)
}

func SaveTeamAttrs(name, resume string, creator int64,email string, sk string) (int64, error) {
	t := &Team{Name: name, Resume: resume, Creator: creator, Email: email, Secretkey: sk,Created: time.Now()}
	return t.Save()
}

func PutUsersInTeam(tid int64, uids string) error {
	if uids == "" {
		return nil
	}

	uidArr := strings.Split(uids, ",")
	for i := 0; i < len(uidArr); i++ {
		uid := uidArr[i]
		if uid == "" {
			continue
		}

		id, err := strconv.Atoi(uid)
		if err != nil {
			return err
		}

		_, err = orm.NewOrm().Raw("insert into rel_team_user(tid,uid) values(?, ?)", tid, id).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func PutAdminInTeam(tid int64, uids string) error {
	if uids == "" {
		return nil
	}

	uidArr := strings.Split(uids, ",")
	for i := 0; i < len(uidArr); i++ {
		uid := uidArr[i]
		if uid == "" {
			continue
		}

		id, err := strconv.Atoi(uid)
		if err != nil {
			return err
		}

        _, err = orm.NewOrm().Raw("insert into rel_team_user(tid,uid,is_admin) values(?, ?, 1)", tid, id).Exec()
		if err != nil {
            _, err = orm.NewOrm().Raw("update rel_team_user set is_admin = 1 where tid=? and uid=?", tid, id).Exec()
            if err != nil {
                return err    
            }
		}
	}

	return nil
}

func (this *Team) UpdateUsers(userIdstr string) error {
	if err := UnlinkByTeamId(this.Id); err != nil {
		return err
	}

	cache.Delete(fmt.Sprintf("t:uids:%d", this.Id))

    uidArr := strings.Split(userIdstr,",")
    for _,uidS := range(uidArr) {
        uId, _ := strconv.ParseInt(uidS,10,64)
        FlushTeamidCache( uId )
    }

	return PutUsersInTeam(this.Id, userIdstr)
}

func (this *Team) UpdateAdmins(userIdstr string) error {
	cache.Delete(fmt.Sprintf("t_admin:uids:%d", this.Id))
    uidArr := strings.Split(userIdstr,",")
    for _,uidS := range(uidArr) {
        uId, _ := strconv.ParseInt(uidS,10,64)
        FlushTeamidCache( uId )
    }

	return PutAdminInTeam(this.Id, userIdstr)
}

func UserIds(tid int64) []int64 {
	key := fmt.Sprintf("t:uids:%d", tid)
	uids := []int64{}
	if err := cache.Get(key, &uids); err != nil {
		uids, err = Uids(tid)
		if err == nil {
			go cache.Set(key, uids, time.Hour)
		}
	}
	return uids
}

func AdminUserIds(tid int64) []int64 {
	key := fmt.Sprintf("t_admin:uids:%d", tid)
	uids := []int64{}
	if err := cache.Get(key, &uids); err != nil {
		uids, err = AdminUids(tid)
		if err == nil {
			go cache.Set(key, uids, time.Hour)
		}
	}
	return uids
}

func MembersByTeamName(name string) []*User {
	ret := []*User{}
	if name == "" {
		return ret
	}

	return MembersByTeamId(ReadTeamIdByName(name))
}

func AdminsByTeamName(name string) []*User {
	ret := []*User{}
	if name == "" {
		return ret
	}

	return AdminsByTeamId(ReadTeamIdByName(name))
}

func MembersByTeamId(tid int64) []*User {
	ret := []*User{}
	if tid <= 0 {
		return ret
	}

	uids := UserIds(tid)
	size := len(uids)
	if size == 0 {
		return ret
	}

	for _, uid := range uids {
		u := ReadUserById(uid)
		if u == nil {
			continue
		}

		ret = append(ret, u)
	}

	return ret
}

func AdminsByTeamId(tid int64) []*User {
	ret := []*User{}
	if tid <= 0 {
		return ret
	}

	uids := AdminUserIds(tid)
	size := len(uids)
	if size == 0 {
		return ret
	}

	for _, uid := range uids {
		u := ReadUserById(uid)
		if u == nil {
			continue
		}

		ret = append(ret, u)
	}

	return ret
}

func (this *Team) Remove() error {
	err := UnlinkByTeamId(this.Id)
	if err != nil {
		return err
	}

	num, err := DeleteTeamById(this.Id)
	if err == nil && num > 0 {
		cache.Delete(fmt.Sprintf("team:id:%s", this.Name))
		cache.Delete(fmt.Sprintf("team:obj:%d", this.Id))
	}

	return err
}

func UnlinkByTeamId(id int64) error {
	uids, err := Uids(id)
	if err == nil && len(uids) > 0 {
		for i := 0; i < len(uids); i++ {
			cache.Delete(fmt.Sprintf("u:tids:%d", uids[i]))
		}
	}
	_, err = orm.NewOrm().Raw("DELETE FROM `rel_team_user` WHERE `tid` = ?", id).Exec()
	return err
}

func UnlinkByUserId(id int64) error {
	_, err := orm.NewOrm().Raw("DELETE FROM `rel_team_user` WHERE `uid` = ?", id).Exec()
	return err
}

func DeleteTeamById(id int64) (int64, error) {
	r, err := orm.NewOrm().Raw("DELETE FROM `team` WHERE `id` = ?", id).Exec()
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (this *Team) UserIds() string {
	uids := UserIds(this.Id)
	size := len(uids)
	if size == 0 {
		return ""
	}

	arr := make([]string, size)
	for i := 0; i < size; i++ {
		arr[i] = fmt.Sprintf("%d", uids[i])
	}

	return strings.Join(arr, ",")
}

func (this *Team) AdminUserIds() string {
    uids := AdminUserIds(this.Id)
	size := len(uids)
	if size == 0 {
		return ""
	}

	arr := make([]string, size)
	for i := 0; i < size; i++ {
		arr[i] = fmt.Sprintf("%d", uids[i])
	}

	return strings.Join(arr, ",")
}

func (this *Team) IsAdmin(uid int64) bool {
    adminUids := this.AdminUserIds()    
	uidArr := strings.Split(adminUids, ",")
    for _,v := range uidArr {
        uInt ,_ := strconv.ParseInt(v, 10, 64)  
        if uInt == uid {
            return true    
        }
    }
    
    return false
}

func (this *Team) Update() (int64, error) {
	num, err := orm.NewOrm().Update(this)

	if err == nil && num > 0 {
		cache.Delete(fmt.Sprintf("team:obj:%d", this.Id))
	}

	return num, err
}
