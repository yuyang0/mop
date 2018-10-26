// Copyright (c) 2013-2016 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
)

const marketURL = `http://qt.gtimg.cn/?q=sh000001,sz399001,sz399006,hkHSI,hkHSCEI,usDJI,usINX,usIXIC`

// Market stores current market information displayed in the top three lines of
// the screen. The market data is fetched and parsed from the HTML page above.
type Market struct {
	IsClosed  bool              // True when U.S. markets are closed.
	ShIndex   map[string]string // 上证指数
	SzIndex   map[string]string // 深圳成指
	CybIndex  map[string]string //创业板指数
	Dow       map[string]string // Hash of Dow Jones indicators.
	Nasdaq    map[string]string // Hash of NASDAQ indicators.
	Sp500     map[string]string // Hash of S&P 500 indicators.
	Tokyo     map[string]string
	HongKong  map[string]string
	London    map[string]string
	Frankfurt map[string]string
	Yield     map[string]string
	Oil       map[string]string
	Yen       map[string]string
	Euro      map[string]string
	Gold      map[string]string
	regex     *regexp.Regexp // Regex to parse market data from HTML.
	errors    string         // Error(s), if any.
}

// Returns new initialized Market struct.
func NewMarket() *Market {
	market := &Market{}
	market.IsClosed = false
	market.ShIndex = make(map[string]string)
	market.SzIndex = make(map[string]string)
	market.CybIndex = make(map[string]string)
	market.Dow = make(map[string]string)
	market.Nasdaq = make(map[string]string)
	market.Sp500 = make(map[string]string)

	market.Tokyo = make(map[string]string)
	market.HongKong = make(map[string]string)
	market.London = make(map[string]string)
	market.Frankfurt = make(map[string]string)

	market.Yield = make(map[string]string)
	market.Oil = make(map[string]string)
	market.Yen = make(map[string]string)
	market.Euro = make(map[string]string)
	market.Gold = make(map[string]string)

	market.errors = ``

	market.regex, _ = regexp.Compile("_\\w+")

	return market
}

// Fetch downloads HTML page from the 'marketURL', parses it, and stores resulting data
// in internal hashes. If download or data parsing fails Fetch populates 'market.errors'.
func (market *Market) Fetch() (self *Market) {
	self = market // <-- This ensures we return correct market after recover() from panic().
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		market.errors = fmt.Sprintf("Error fetching market data...\n%s", err)
	// 	}
	// }()

	response, err := http.Get(marketURL)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	body = market.isMarketOpen(body)
	// return market.extract(body)
	return market.extract(market.trim(body))
}

// Ok returns two values: 1) boolean indicating whether the error has occured,
// and 2) the error text itself.
func (market *Market) Ok() (bool, string) {
	return market.errors == ``, market.errors
}

//-----------------------------------------------------------------------------
func (market *Market) isMarketOpen(body []byte) []byte {
	// TBD -- CNN page doesn't seem to have market open/close indicator.
	return body
}

//-----------------------------------------------------------------------------
func (market *Market) trim(body []byte) []byte {
	return bytes.Replace(bytes.TrimSpace(body), []byte{'\n'}, []byte{}, -1)
}

//-----------------------------------------------------------------------------
func (market *Market) extract(body []byte) *Market {
	lines := bytes.Split(body, []byte{';'})

	for _, line := range lines {
		if len(line) == 0 {
			break
		}
		stockCode := market.regex.Find(line)
		if len(stockCode) > 0 {
			stockCode = stockCode[1:]
		}
		columns := bytes.Split(bytes.TrimSpace(line), []byte{'~'})
		if len(columns) == 0 {
			break
		}
		switch string(stockCode) {
		case "sh000001":
			market.ShIndex[`change`] = string(columns[31])
			market.ShIndex[`latest`] = string(columns[3])
			market.ShIndex[`percent`] = string(columns[32])
		case "sz399001":
			market.SzIndex[`change`] = string(columns[31])
			market.SzIndex[`latest`] = string(columns[3])
			market.SzIndex[`percent`] = string(columns[32])
		case "sz399006":
			market.CybIndex[`change`] = string(columns[31])
			market.CybIndex[`latest`] = string(columns[3])
			market.CybIndex[`percent`] = string(columns[32])
		case "hkHSI":
			market.HongKong[`change`] = string(columns[31])
			market.HongKong[`latest`] = string(columns[3])
			market.HongKong[`percent`] = string(columns[32])
		case "usDJI":
			market.Dow[`change`] = string(columns[31])
			market.Dow[`latest`] = string(columns[3])
			market.Dow[`percent`] = string(columns[32])
		case "usINX":
			market.Sp500[`change`] = string(columns[31])
			market.Sp500[`latest`] = string(columns[3])
			market.Sp500[`percent`] = string(columns[32])
		case "usIXIC":
			market.Nasdaq[`change`] = string(columns[31])
			market.Nasdaq[`latest`] = string(columns[3])
			market.Nasdaq[`percent`] = string(columns[32])
		}
	}

	return market
}
