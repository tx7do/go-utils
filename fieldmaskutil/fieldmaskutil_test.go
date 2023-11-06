package fieldmaskutil

import (
	"reflect"
	"testing"
)

func Test_NestedMaskFromPaths(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want NestedMask
	}{
		{
			name: "no nested fields",
			args: args{paths: []string{"a", "b", "c"}},
			want: NestedMask{"a": NestedMask{}, "b": NestedMask{}, "c": NestedMask{}},
		},
		{
			name: "with nested fields",
			args: args{paths: []string{"aaa.bb.c", "dd.e", "f"}},
			want: NestedMask{
				"aaa": NestedMask{"bb": NestedMask{"c": NestedMask{}}},
				"dd":  NestedMask{"e": NestedMask{}},
				"f":   NestedMask{}},
		},
		{
			name: "single field",
			args: args{paths: []string{"a"}},
			want: NestedMask{"a": NestedMask{}},
		},
		{
			name: "empty fields",
			args: args{paths: []string{}},
			want: NestedMask{},
		},
		{
			name: "invalid input",
			args: args{paths: []string{".", "..", "..."}},
			want: NestedMask{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NestedMaskFromPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NestedMaskFromPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkNestedMaskFromPaths(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NestedMaskFromPaths([]string{"aaa.bbb.c.d.e.f", "aa.b.cc.ddddddd", "e", "f", "g.h.i.j.k"})
	}
}
