package mop

import (
	"testing"
)

func Test_Division_1(t *testing.T) {
	profile := NewProfile()
	market := NewMarket()
	quotes := NewQuotes(market, profile)
	quotes.Fetch()
	// t.Log("第一个测试通过了") //记录一些你期望记录的信息
}

// func Test_Division_2(t *testing.T) {
// 	t.Error("就是不通过")
// }
