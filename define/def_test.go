/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: def.go
 * @time: 2019/8/5 13:48
 */
package define

import (
	`fmt`
	"testing"
)

func TestValidateGwType(t *testing.T) {
	fmt.Println(SerializeError, EncryptError, DecryptError, CompressError)
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "TestValidateGwType",
			args: args{GwTypeUdp},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateGwType(tt.args.in); got != tt.want {
				t.Errorf("ValidateGwType() = %v, want %v", got, tt.want)
			}
		})
	}
}
