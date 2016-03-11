// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ping

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"github.com/fcavani/e"
	"github.com/xordataexchange/crypt/config"
	"gopkg.in/fcavani/gormethods.v3/core"
	"gopkg.in/yaml.v2"
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
	for i := 0; i < 10; i++ {
		err = c.Ping()
		if err != nil {
			return e.Forward(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

var EtcdSecRing = ""

type etcd struct {
	Endpoints  []string
	SecKeyRing string
	kapi       client.KeysAPI
	cm         config.ConfigManager
}

func (etc *etcd) Init() error {
	var err error
	if len(etc.Endpoints) == 0 {
		return e.New("no end points")
	}
	cfg := client.Config{
		Endpoints: etc.Endpoints,
		//Transport: http.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		return e.Forward(err)
	}
	etc.kapi = client.NewKeysAPI(c)

	if etc.SecKeyRing == "" {
		return nil
	}

	kr, err := os.Open(etc.SecKeyRing)
	if err != nil {
		return e.Forward(err)
	}
	defer kr.Close()
	etc.cm, err = config.NewEtcdConfigManager(etc.Endpoints, kr)
	if err != nil {
		return e.Forward(err)
	}

	return nil
}

func (etc *etcd) Get(key string, opt *client.GetOptions) ([]byte, error) {
	var err error
	if etc.SecKeyRing == "" {
		if opt == nil {
			opt = &client.GetOptions{}
		}
		resp, err := etc.kapi.Get(context.Background(), key, opt)
		if err != nil {
			return nil, e.Forward(err)
		}
		if len(resp.Node.Nodes) > 0 {
			fmt.Println(resp.Node.Nodes)
		}
		return []byte(resp.Node.String()), nil
	}
	buf, err := etc.cm.Get(key)
	if err != nil {
		return nil, e.Forward(err)
	}
	return buf, nil
}

func (etc *etcd) TryGetUntil(key string, timeout time.Duration) (buf []byte, err error) {
	op := func() (err error) {
		buf, err = etc.Get(key, nil)
		if err != nil {
			err = e.Forward(err)
			return
		}
		return
	}
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	err = backoff.Retry(op, backOff)
	if err != nil {
		err = e.Forward(err)
		return
	}
	return
}

type Instance struct {
	Name     string
	Session  string
	Instance string
	Pid      string
	Proto    string
	Addr     string
	Hostname string
}

func NewInstance(in []byte) (*Instance, error) {
	inst := new(Instance)
	err := yaml.Unmarshal(in, inst)
	if err != nil {
		return nil, e.Forward(err)
	}
	return inst, nil
}

func getInst(etc *etcd, path string) (*Instance, error) {
	buf, err := etc.TryGetUntil(
		path,
		2*time.Minute,
	)
	if err != nil {
		return nil, e.Forward(err)
	}
	i, err := NewInstance([]byte(buf))
	if err != nil {
		return nil, e.Forward(err)
	}
	return i, nil
}

// PingGormethodsEtcd uses etcd to retrieve the session and instance names.
// Url format: gormethods+etc://id:pass@etcd_endpoint/path/to/etcd/data
// The data must be in the format of Instace struct.
// To cryptograph the data set the variable EtcdSecRing to the location of the
// pgp key.
func PingGormethodsEtcd(u *url.URL) error {
	etc := &etcd{
		Endpoints:  []string{u.Host},
		SecKeyRing: EtcdSecRing,
	}
	inst, err := getInst(etc, u.Path)
	if err != nil {
		return e.Forward(err)
	}
	u.Scheme = "gormethods"
	u.Host = inst.Hostname
	u.Path = "/" + inst.Session + "/" + inst.Instance
	return e.Forward(PingGormethods(u))
}

func init() {
	Add("gormethods", PingGormethods)
}
