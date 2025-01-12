package paramstore

import (
	"reflect"
	"testing"
)

func Test_ToSliceString(t *testing.T) {
	tests := map[string]struct {
		parameters Parameters
		want       []string
	}{
		"parameters convert to slice string": {
			parameters: validTestdata.toParameters(),
			want:       validTestdata.toSliceString(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.parameters.ToSliceString()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf(
					"ToSliceString() returns mismatched slices; want=%+v, got=%+v\n",
					tt.want,
					got,
				)
			}
		})
	}
}
