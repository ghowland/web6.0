package web6

import (
	"github.com/ghowland/yudien/yudien"
	"github.com/ghowland/yudien/yudiendata"
)

type Web6Config struct {
	//Ldap  yudien.LdapConfig  `json:"ldap"`
	DefaultDatabase yudiendata.DatabaseConfig            `json:"default_database"`
	Databases       map[string]yudiendata.DatabaseConfig `json:"databases"`
	LdapOverride    yudiendata.DatabaseConfig            `json:"ldap_override"`
	Authentication  yudien.AuthenticationConfig          `json:"authentication"`
	Logging         yudien.LoggingConfig                 `json:"logging"`
}
