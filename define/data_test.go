/**
 * @version: 1.0.0
 * @author: zgd: general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: data.go
 * @time: 2019-08-13 12:45
 */

package define

import (
	"testing"
)

func Test_is2n(t *testing.T) {
	type args struct {
		v uint32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "is2n_1",
			args: args{
				v: 8,
			},
			want: true,
		},
		{
			name: "is2n_2",
			args: args{
				v: 10,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := is2n(tt.args.v); got != tt.want {
				t.Errorf("is2n() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkCenter_filterMobilePlatform(t *testing.T) {

	type args struct {
		platform uint32
	}
	tests := []struct {
		name   string
		args   args
		want   uint32
	}{
		// TODO: Add test cases.
		{
			name:"TestLinkCenter_filterMobilePlatform_1",
			args:args{
				platform:1,
			},
			want:1,
		},
		{
			name:"TestLinkCenter_filterMobilePlatform_2",
			args:args{
				platform:2,
			},
			want:2,
		},
		{
			name:"TestLinkCenter_filterMobilePlatform_3",
			args:args{
				platform:3,
			},
			want:3,
		},
		{
			name:"TestLinkCenter_filterMobilePlatform_3",
			args:args{
				platform:5,
			},
			want:3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLinkCenter()
			if got := p.filterMobilePlatform(tt.args.platform); got != tt.want {
				t.Errorf("LinkCenter.filterMobilePlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}
