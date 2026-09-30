package main

import (
	"reflect"
	"testing"
)

func TestWithOwnDir(t *testing.T) {
	for _, c := range []struct {
		name string
		env  []string
		exe  string
		want []string
	}{
		{"missing from PATH", []string{"HOME=/h", "PATH=/usr/bin:/bin"}, "/h/go/bin/backdrop-shell",
			[]string{"HOME=/h", "PATH=/usr/bin:/bin:/h/go/bin"}},
		{"already there", []string{"PATH=/h/go/bin/:/usr/bin"}, "/h/go/bin/backdrop-shell",
			[]string{"PATH=/h/go/bin/:/usr/bin"}},
		{"empty PATH", []string{"PATH="}, "/h/go/bin/backdrop-shell", []string{"PATH=/h/go/bin"}},
		{"no PATH", []string{"HOME=/h"}, "/h/go/bin/backdrop-shell", []string{"HOME=/h", "PATH=/h/go/bin"}},
		{"unknown binary", []string{"PATH=/usr/bin"}, "", []string{"PATH=/usr/bin"}},
	} {
		env := append([]string(nil), c.env...)
		if got := withOwnDir(env, c.exe); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
		if !reflect.DeepEqual(env, c.env) {
			t.Errorf("%s: the environment passed in was changed to %q", c.name, env)
		}
	}
}
