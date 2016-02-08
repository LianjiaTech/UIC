package uic

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/toolkits/cache"
	"github.com/toolkits/logger"
	"time"
)

func SelectSessionBySig(sig string) *Session {
	if sig == "" {
		return nil
	}

	obj := Session{Sig: sig}
	err := orm.NewOrm().Read(&obj, "Sig")
	if err != nil {
		if err != orm.ErrNoRows {
			logger.Errorln(err)
		}
		return nil
	}
	return &obj
}

func ReadSessionBySig(sig string) *Session {
	if sig == "" {
		return nil
	}

	key := fmt.Sprintf("session:obj:%s", sig)
	var obj Session
	if err := cache.Get(key, &obj); err != nil {
		objPtr := SelectSessionBySig(sig)
		if objPtr != nil {
			go cache.Set(key, objPtr, time.Hour)
		}
		return objPtr
	}

	return &obj
}

func (this *Session) Save() (int64, error) {
	return orm.NewOrm().Insert(this)
}

func SaveSessionAttrs(uid int64, sig string, expired int) (int64, error) {
	s := &Session{Uid: uid, Sig: sig, Expired: expired}
	return s.Save()
}

func RemoveSessionByUid(uid int64) {
	var ss []Session
	Sessions().Filter("Uid", uid).All(&ss, "Id", "Sig")
	if ss == nil || len(ss) == 0 {
		return
	}

	for _, s := range ss {
		num, err := DeleteSessionById(s.Id)
		if err == nil && num > 0 {
			cache.Delete(fmt.Sprintf("session:obj:%s", s.Sig))
		}
	}
}

func DeleteSessionById(id int64) (int64, error) {
	r, err := orm.NewOrm().Raw("DELETE FROM `session` WHERE `id` = ?", id).Exec()
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func Sessions() orm.QuerySeter {
	return orm.NewOrm().QueryTable(new(Session))
}
