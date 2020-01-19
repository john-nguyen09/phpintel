// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import "context"

type contextKey int

const (
	clientKey     = contextKey(iota)
	versionKey    = contextKey(iota)
	memprofileKey = contextKey(iota)
)

func WithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey, version)
}

func WithMemprofile(ctx context.Context, memprofile string) context.Context {
	return context.WithValue(ctx, memprofileKey, memprofile)
}

func GetVersion(ctx context.Context) string {
	value := ctx.Value(versionKey)
	if version, ok := value.(string); ok {
		return version
	}
	return ""
}

func GetMemprofile(ctx context.Context) string {
	value := ctx.Value(memprofileKey)
	if memprofile, ok := value.(string); ok {
		return memprofile
	}
	return ""
}
