/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: fun.go
 * @time: 2019/8/6 21:11
 */
package util

import (
	`github.com/astaxie/beego/logs`
	`github.com/funny/slab`
)

func MakePool(t string, minChunk, maxChunk, factor, pageSize int) slab.Pool {
	switch t {
	case "sync":
		return slab.NewSyncPool(minChunk, maxChunk, factor)
	case "atom":
		return slab.NewAtomPool(minChunk, maxChunk, factor, pageSize)
	case "chan":
		return slab.NewChanPool(minChunk, maxChunk, factor, pageSize)
	default:
		logs.Error(`unsupported memory pool type, must be "sync", "atom" or "chan"`)
	}
	return nil
}



