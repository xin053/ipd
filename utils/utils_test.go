package utils

import (
	"reflect"
	"testing"

	"github.com/xin053/ipd/config"
)

func TestEmptyStrings(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"test1",
			args{[]string{"", "123"}},
			false,
		},
		{
			"test2",
			args{[]string{"", "", "", ""}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmptyStrings(tt.args.s...); got != tt.want {
				t.Errorf("EmptyStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddGeo(t *testing.T) {
	type args struct {
		ipInfo *config.IPInfo
	}
	tests := []struct {
		name string
		args args
		want config.IPWithGeo
	}{
		{
			"test1",
			args{&config.IPInfo{
				"183.238.58.120",
				"中国",
				"广东",
				"",
				"",
			}},
			config.IPWithGeo{
				config.IPInfo{
					"183.238.58.120",
					"中国",
					"广东",
					"",
					"",
				},
				config.Geo{
					113.23,
					23.16,
				},
			},
		},
		{
			"test2",
			args{&config.IPInfo{
				"183.238.58.120",
				"",
				"",
				"广州",
				"",
			}},
			config.IPWithGeo{
				config.IPInfo{
					"183.238.58.120",
					"中国",
					"",
					"广州",
					"",
				},
				config.Geo{
					113.23,
					23.16,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddGeo(tt.args.ipInfo); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("AddGeo() = %v, want %v", *got, tt.want)
			}
		})
	}
}
