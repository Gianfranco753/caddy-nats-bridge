package caddynats

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"

	"github.com/caddyserver/caddy/v2"
)

// Returns a basic request for testing
func reqPath(path string) *http.Request {
	req, _ := http.NewRequest("GET", "http://localhost"+path, &bytes.Buffer{})
	return req
}

func TestAddNatsPublishVarsToReplacer(t *testing.T) {
	type test struct {
		req *http.Request

		input string
		want  string
	}

	tests := []test{

		// Basic subject mapping
		{req: reqPath("/foo/bar"), input: "{nats.subject}", want: "foo.bar"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject}", want: "foo.bar.bat.baz"},
		{req: reqPath("/foo/bar/bat/baz?query=true"), input: "{nats.subject}", want: "foo.bar.bat.baz"},
		{req: reqPath("/foo/bar"), input: "prefix.{nats.subject}.suffix", want: "prefix.foo.bar.suffix"},

		// Segment placeholders
		{req: reqPath("/foo/bar"), input: "{nats.subject.0}", want: "foo"},
		{req: reqPath("/foo/bar"), input: "{nats.subject.1}", want: "bar"},

		// Segment Ranges
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.0:}", want: "foo.bar.bat.baz"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.1:}", want: "bar.bat.baz"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.2:}", want: "bat.baz"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.1:2}", want: "bar.bat"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.0:2}", want: "foo.bar.bat"},
		{req: reqPath("/foo/bar/bat/baz"), input: "{nats.subject.:2}", want: "foo.bar.bat"},

		// TODO Prefix?
	}

	for _, tc := range tests {
		repl := caddy.NewReplacer()
		addNATSPublishVarsToReplacer(repl, tc.req, nil, "")
		got := repl.ReplaceAll(tc.input, "")
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("expected: %v, got: %v", tc.want, got)
		}
	}
}