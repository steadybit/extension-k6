package extk6

import (
	"github.com/steadybit/extension-kit/extutil"
	"reflect"
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
		{name: "match", args: args{value: "abc", after: "b"}, want: extutil.Ptr("c")},
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
		{name: "match", args: args{lines: []string{"abc", "def", "level=error msg=something"}}, want: extutil.Ptr("something")},
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
