/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: calculate.go
 * @time: 2019/8/21 20:40
 */
package qps

import (
	"fmt"
	"sync/atomic"

	"github.com/generalzgd/grpc-gw/statistic"
)

func init() {
	statistic.Register(&StatQps{
		threshold: 5000, // todo 阀值可能需要视情况来定，这个100是一个想象值
	})
}

type StatQps struct {
	procNum int32
	//procDur     time.Duration
	threshold   int
	oldDebugQps int
}

func (p *StatQps) NewRecord(typ int, args ...interface{}) {
	if typ != statistic.StatQps {
		return
	}
	p.procNum++
	//for _, arg := range args {
	//if d, ok := arg.(time.Duration); ok {
	//p.procDur += d
	//}
	//break
	//}
}

func (p *StatQps) GetType() int {
	return statistic.Stat_Qps
}

func (p *StatQps) GetCount() int {
	return int(atomic.SwapInt32(&p.procNum, 0))
	//qpsamt := p.procDur/time.Second
	//if qpsamt > 0 {
	//	v := int(float64(p.procNum) / qpsamt)
	// 重置数据
	//p.procDur = 0
	//p.procNum = 0
	//return v
	//}
	//return 0
}

func (p *StatQps) GetInterval() int {
	return 30 // 30秒
}

func (p *StatQps) GetThreshold() int {
	return p.threshold
}

func (p *StatQps) MakeWarnStr(num int, threshold int) (string, bool) {
	if num > 0 && threshold > 0 && num > threshold {
		return fmt.Sprintf("告警 Qps:%d, threshold:%d", num/p.GetInterval(), threshold), true
	}
	return "", false
}

func (p *StatQps) Debug(qps int, threshold int) (string, bool) {
	if qps > 0 && qps != p.oldDebugQps {
		p.oldDebugQps = qps // 防止日志刷屏
		return fmt.Sprintf("Qps:%d, threshold:%d", qps, threshold), true
	}
	return "", false
}
