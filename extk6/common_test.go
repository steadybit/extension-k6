package extk6

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func Test_substringAfter(t *testing.T) {
	type args struct {
		value string
		after string
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{name: "match", args: args{value: "abc", after: "b"}, want: new("c")},
		{name: "no match", args: args{value: "aaa", after: "b"}, want: nil},
		{name: "no match after last", args: args{value: "abc", after: "c"}, want: nil},
		{name: "empty", args: args{value: "", after: "b"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := substringAfter(tt.args.value, tt.args.after); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("substringAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractErrorFromStdOut(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{name: "match", args: args{lines: []string{"abc", "def", "level=error msg=something"}}, want: new("something")},
		{name: "no match", args: args{lines: []string{"abc", "def", "level=debug msg=something"}}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractErrorFromStdOut(tt.args.lines); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractErrorFromStdOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_k6Environment(t *testing.T) {
	t.Setenv("STEADYBIT_EXTENSION_CLOUD_API_TOKEN", "secret")
	t.Setenv("SOME_OTHER_SECRET", "secret")
	t.Setenv("K6_PROMETHEUS_RW_SERVER_URL", "http://prometheus:9090/api/v1/write")
	t.Setenv("HTTPS_PROXY", "http://proxy:3128")

	env := k6Environment()

	for _, want := range []string{"K6_PROMETHEUS_RW_SERVER_URL=http://prometheus:9090/api/v1/write", "HTTPS_PROXY=http://proxy:3128"} {
		if !slices.Contains(env, want) {
			t.Errorf("k6Environment() is missing %q", want)
		}
	}
	for _, entry := range env {
		if strings.HasPrefix(entry, "STEADYBIT_EXTENSION_CLOUD_API_TOKEN=") || strings.HasPrefix(entry, "SOME_OTHER_SECRET=") {
			t.Errorf("k6Environment() leaks %q", entry)
		}
	}
}
