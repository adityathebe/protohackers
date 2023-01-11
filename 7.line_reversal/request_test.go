package main

import (
	"reflect"
	"testing"
)

func Test_validateMsg(t *testing.T) {
	tests := []struct {
		name string
		args string
		want Msg
	}{
		{args: "/connect/1234567/", want: Msg{mType: "connect", sessID: 1234567}},
		{args: "/data/1234567/0/hello/", want: Msg{mType: "data", sessID: 1234567, pos: 0, data: "hello"}},
		{args: "/ack/1234567/5/", want: Msg{mType: "ack", sessID: 1234567, ackLen: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := parseMsg(tt.args); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("parseMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_escapeData(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{args: `foo/bar\baz`, want: `foo\/bar\\baz`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeData(tt.args); got != tt.want {
				t.Errorf("escapeData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unescapeData(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{want: `foo/bar\baz`, args: `foo\/bar\\baz`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unescapeData(tt.args); got != tt.want {
				t.Errorf("unescapeData() = %v, want %v", got, tt.want)
			}
		})
	}
}
