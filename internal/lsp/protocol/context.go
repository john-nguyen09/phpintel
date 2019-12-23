// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import "context"

type contextKey int

const (
	clientKey     = contextKey(iota)
	cpuprofileKey = contextKey(iota)
	versionKey    = contextKey(iota)
)

func WithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

func WithCpuProfile(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, cpuprofileKey, value)
}

func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey, version)
}

func HasCpuProfile(ctx context.Context) bool {
	value := ctx.Value(cpuprofileKey)
	if cpuprofile, ok := value.(bool); ok {
		return cpuprofile
	}
	return false
}

func GetVersion(ctx context.Context) string {
	value := ctx.Value(versionKey)
	if version, ok := value.(string); ok {
		return version
	}
	return ""
}
