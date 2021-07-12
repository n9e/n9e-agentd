package mysql

import (
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

func (c *Check) getVersion() error {
	v := &c.version
	_, err := c.queryRow("SELECT VERSION()", &v.rawVersion)
	if err != nil {
		return err
	}

	// Version might include a build, a flavor, or both
	// e.g. 4.1.26-log, 4.1.26-MariaDB, 10.0.1-MariaDB-mariadb1precise-log
	// See http://dev.mysql.com/doc/refman/4.1/en/information-functions.html#function_version
	// https://mariadb.com/kb/en/library/version/
	// and https://mariadb.com/kb/en/library/server-system-variables/#version
	parts := strings.Split(v.rawVersion, "-")
	v.version = parts[0]

	for _, data := range parts {
		if data == "MariaDB" {
			v.flavor = "MariaDB"
		}
		if data != "MariaDB" && v.flavor == "" {
			v.flavor = "MySQL"
		}
		for _, build := range BUILDS {
			if data == build {
				v.build = build
			}
		}
		if v.build == "" {
			v.build = "unspecified"
		}
	}

	return nil
}

type MySQLVersion struct {
	rawVersion string
	version    string
	flavor     string
	build      string
}

func (p *MySQLVersion) versionCompatible(compatVersion ...int) bool {
	if len(compatVersion) < 3 {
		compatVersion = append(compatVersion, 0, 0, 0)
	}
	mysqlVersion := strings.Split(p.version, ".")
	klog.V(5).Infof("MySQL version %v", mysqlVersion)

	for i := 0; i < len(mysqlVersion[2]); i++ {
		c := mysqlVersion[2][i]
		if c < '0' || c > '9' {
			mysqlVersion[2] = mysqlVersion[2][:i]
			break
		}
	}

	for i := 0; i < 2; i++ {
		if n, _ := strconv.Atoi(mysqlVersion[i]); n < compatVersion[i] {
			return false
		}
	}

	return true
}
