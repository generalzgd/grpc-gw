/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: istatistic.go
 * @time: 2019/8/22 8:53
 */
package statistic

type IMonitor interface {
	NewRecord(int, ...interface{})
	GetType() int
	// 获取数据的时候，同时将已有的数据设为0
	GetCount() int
	GetInterval() int // 秒
	GetThreshold() int
	MakeWarnStr(int, int) (string, bool)
	Debug(int, int)(string,bool)
}
