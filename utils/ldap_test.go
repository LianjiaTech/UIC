package utils

import (
	"log"
	"testing"
)

const (
	addr       = "ldap.example.com:3268"
	baseDN     = "dc=example,dc=com"
	bindDN     = "cn=Manager,dc=example,dc=com"
	bindPasswd = "12345678"
	user       = "test"
	password   = "12345678"
	UserField  = "uid"
)

var Attributes []string = []string{"cn", "mail", "telephoneNumber"}

func Test_ldap_bind_fe(t *testing.T) {
	sucess, err := LdapBind(addr, baseDN, bindDN, bindPasswd, UserField, user, password)
	log.Println("sucess:", sucess)
	log.Println("err", err)
}
func Test_ldap_search_fe(t *testing.T) {
	user_attributes, err := Ldapsearch(addr, baseDN, bindDN, bindPasswd, UserField, user, Attributes)
	log.Println("user_attributes", user_attributes)
	log.Println("err", err)
}
