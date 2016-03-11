// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ping

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/fcavani/e"
	"gopkg.in/fcavani/gormethods.v3/auth"
	"gopkg.in/fcavani/gormethods.v3/core"
)

type DummyServer struct{}

func (d *DummyServer) Foo() {}

func startServer() (*core.Server, error) {
	instances := core.NewInstances()

	owners := auth.NewPlainTextCredentials()
	id := &auth.PlainText{Ident: "id", Pass: "pass1234"}
	owners.Add(id)

	constructor := core.NewConstructor()

	server := &core.Server{
		Proto:       "tcp",
		Addr:        "localhost:0",
		ConnTimeout: 30 * time.Second,
		Insts:       instances,
		Cons:        constructor,
		FifoLength:  1,
		Ttl:         24 * time.Hour,
		Cleanup:     5 * time.Minute,
	}
	err := server.Start()
	if err != nil {
		return nil, e.Forward(err)
	}

	err = instances.New("1", "1", reflect.ValueOf(&DummyServer{}), owners)
	if err != nil {
		return nil, e.Forward(err)
	}
	return server, nil
}

func TestPingGormethods(t *testing.T) {
	server, err := startServer()
	if err != nil {
		t.Fatal(err)
	}

	gormethodsURL := "gormethods://id:pass1234@" + server.Address().String() + "/1/1"

	u, err := url.Parse(gormethodsURL)
	if err != nil {
		t.Fatal(err)
	}

	err = PingGormethods(u)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}
