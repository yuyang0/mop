package mop

import (
	"testing"
)

func Test_Maeket(t *testing.T) {
	market := NewMarket()
	market.Fetch()
	// t.Log("第一个测试通过了") //记录一些你期望记录的信息
}

// func Test_Division_2(t *testing.T) {
// 	t.Error("就是不通过")
// }
