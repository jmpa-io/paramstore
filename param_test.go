package paramstore

import (
	"context"
	"reflect"
	"testing"
)

func Test_NamesToSliceString(t *testing.T) {
	tests := map[string]struct {
		params Params
		want   []string
	}{
		"param names convert to slice string": {
			params: Params{
				{Name: "hello", Value: "this is ignored"},
				{Name: "world", Value: "xxxx"},
			},
			want: []string{
				"hello",
				"world",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.params.NamesToSliceString(context.Background())
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NamesToSliceString() returns mismatched slices; want=%+v, got=%+v\n", tt.want, got)
			}
		})
	}
}
