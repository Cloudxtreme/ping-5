// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ping

import (
	"database/sql"
	"net"
	"net/url"

	"github.com/fcavani/e"
	"github.com/fcavani/log"
	"github.com/go-sql-driver/mysql"
)

func init() {
	mysql.RegisterDial("tcp", func(addr string) (net.Conn, error) {
		return (&net.Dialer{Timeout: DialTimeout}).Dial("tcp", addr)
	})
	mysql.RegisterDial("unix", func(addr string) (net.Conn, error) {
		return (&net.Dialer{Timeout: DialTimeout}).Dial("unix", addr)
	})
	logger := log.Log.Tag("mysql").DebugLevel()
	mysql.SetLogger(logger)
}

// PingMySql connects a mysql server and send a ping.
func PingMySql(u *url.URL) error {
	user := u.User.Username()
	pass, ok := u.User.Password()
	if ok {
		user += ":" + url.QueryEscape(pass)
	}
	uri := user + "@" + u.Host + u.Path
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return e.Forward(err)
	}
	defer db.Close()
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err != nil {
			return e.New(err)
		}
	}
	return nil
}

func init() {
	Add("mysql", PingMySql)
}
