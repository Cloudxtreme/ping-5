// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ping

import (
	"net/url"

	"github.com/fcavani/e"
	"gopkg.in/fcavani/gormethods.v3/core"
)

func PingGormethods(u *url.URL) error {
	c, err := core.NewClientFromUrl(u)
	if err != nil {
		return e.Forward(err)
	}
	err = c.Init()
	if err != nil {
		return e.Forward(err)
	}
	// for i := 0; i < 10; i++ {
	err = c.Ping()
	if err != nil {
		return e.Forward(err)
	}
	// 	time.Sleep(100 * time.Millisecond)
	// }
	return nil
}

func init() {
	Add("gormethods", PingGormethods)
}
