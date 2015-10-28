// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ping

import (
	"os"
)

func OnTravis() bool {
	dir := os.Getenv("TRAVIS_BUILD_DIR")
	return dir != ""
}
