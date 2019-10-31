// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"fmt"
)

// Error takes a message and deliver to logger
func Error(ctx context.Context, message string, err error) {
	if err == nil {
		err = errorString(message)
		message = ""
	}
	// TODO: Implement a real logger
	fmt.Println(err)
}

type errorString string

// Error allows errorString to conform to the error interface.
func (err errorString) Error() string { return string(err) }
