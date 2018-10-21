// Copyright (c) 2013-2016 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// See http://www.gummy-stuff.org/Yahoo-stocks.htm
//
// const quotesURL = `http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=sl1c6k2oghjkva2r2rdyj3j1`
// c2: realtime change vs c1: change
// k2: realtime change vs p2: change
//
const quotesURL = `http://qt.gtimg.cn/q=%s`

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Stock stores quote information for the particular stock ticker. The data
// for all the fields except 'Advancing' is fetched using Yahoo market API.
type Stock struct {
	Name       string // stock name
	Ticker     string // Stock ticker.
	LastTrade  string // l1: last trade.
	Change     string // c6: change real time.
	ChangePct  string // k2: percent change real time.
	Open       string // o: market open price.
	Low        string // g: day's low.
	High       string // h: day's high.
	Low52      string // j: 52-weeks low.
	High52     string // k: 52-weeks high.
	Volume     string // v: volume.
	AvgVolume  string // a2: average volume.
	Turnover   string
	PeRatio    string // r2: P/E ration real time.
	PeRatioX   string // r: P/E ration (fallback when real time is N/A).
	Dividend   string // d: dividend.
	Yield      string // y: dividend yield.
	MarketCap  string // j3: market cap real time.
	MarketCapX string // j1: market cap (fallback when real time is N/A).
	Advancing  bool   // True when change is >= $0.
}

// Quotes stores relevant pointers as well as the array of stock quotes for
// the tickers we are tracking.
type Quotes struct {
	market  *Market  // Pointer to Market.
	profile *Profile // Pointer to Profile.
	stocks  []Stock  // Array of stock quote data.
	re      *regexp.Regexp
	errors  string // Error string if any.
}

// Sets the initial values and returns new Quotes struct.
func NewQuotes(market *Market, profile *Profile) *Quotes {
	re, err := regexp.Compile("_\\w+")
	if err != nil {
		//TODO
		fmt.Print("error when create regex")
	}
	return &Quotes{
		market:  market,
		profile: profile,
		re:      re,
		errors:  ``,
	}
}

// Fetch the latest stock quotes and parse raw fetched data into array of
// []Stock structs.
func (quotes *Quotes) Fetch() (self *Quotes) {
	self = quotes // <-- This ensures we return correct quotes after recover() from panic().
	if quotes.isReady() {
		// defer func() {
		// 	if err := recover(); err != nil {
		// 		quotes.errors = fmt.Sprintf("\n\n\n\nError fetching stock quotes...\n%s", err)
		// 	}
		// }()

		// url := fmt.Sprintf(quotesURL, strings.Join(quotes.profile.Tickers, `,`))
		url := fmt.Sprintf(quotesURL, "sh601225,sh600188")
		response, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		quotes.parse(sanitize(body))
	}

	return quotes
}

// Ok returns two values: 1) boolean indicating whether the error has occured,
// and 2) the error text itself.
func (quotes *Quotes) Ok() (bool, string) {
	return quotes.errors == ``, quotes.errors
}

// AddTickers saves the list of tickers and refreshes the stock data if new
// tickers have been added. The function gets called from the line editor
// when user adds new stock tickers.
func (quotes *Quotes) AddTickers(tickers []string) (added int, err error) {
	if added, err = quotes.profile.AddTickers(tickers); err == nil && added > 0 {
		quotes.stocks = nil // Force fetch.
	}
	return
}

// RemoveTickers saves the list of tickers and refreshes the stock data if some
// tickers have been removed. The function gets called from the line editor
// when user removes existing stock tickers.
func (quotes *Quotes) RemoveTickers(tickers []string) (removed int, err error) {
	if removed, err = quotes.profile.RemoveTickers(tickers); err == nil && removed > 0 {
		quotes.stocks = nil // Force fetch.
	}
	return
}

// isReady returns true if we haven't fetched the quotes yet *or* the stock
// market is still open and we might want to grab the latest quotes. In both
// cases we make sure the list of requested tickers is not empty.
func (quotes *Quotes) isReady() bool {
	return (quotes.stocks == nil || !quotes.market.IsClosed) && len(quotes.profile.Tickers) > 0
}

// Use reflection to parse and assign the quotes data fetched using the Yahoo
// market API.
func (quotes *Quotes) parse(body []byte) *Quotes {
	lines := bytes.Split(body, []byte{';'})
	quotes.stocks = make([]Stock, len(lines))
	//
	// Get the total number of fields in the Stock struct. Skip the last
	// Advanicing field which is not fetched.
	//
	// fieldsCount := reflect.ValueOf(quotes.stocks[0]).NumField() - 1
	//
	// Split each line into columns, then iterate over the Stock struct
	// fields to assign column values.
	//
	for i, line := range lines {
		if len(line) == 0 {
			break
		}
		stockCode := quotes.re.Find(line)
		if len(stockCode) > 0 {
			stockCode = stockCode[1:]
		}
		columns := bytes.Split(bytes.TrimSpace(line), []byte{'~'})
		stockName, _ := GbkToUtf8(columns[1])
		quotes.stocks[i] = Stock{
			Name:       string(stockName),
			Ticker:     string(stockCode),
			LastTrade:  string(columns[3]),
			Change:     string(columns[31]),
			ChangePct:  string(columns[32]),
			Open:       string(columns[5]),
			Low:        string(columns[34]),
			High:       string(columns[33]),
			Low52:      "N/A",
			High52:     "N/A",
			Volume:     string(columns[36]),
			AvgVolume:  "N/A",
			Turnover:   string(columns[37]),
			PeRatio:    string(columns[39]),
			PeRatioX:   "N/A",
			Dividend:   "N/A",
			Yield:      "N/A",
			MarketCap:  string(columns[45]),
			MarketCapX: "N/A",
		}
		//
		// Try realtime value and revert to the last known if the
		// realtime is not available.
		//
		if quotes.stocks[i].PeRatio == `N/A` && quotes.stocks[i].PeRatioX != `N/A` {
			quotes.stocks[i].PeRatio = quotes.stocks[i].PeRatioX
		}
		if quotes.stocks[i].MarketCap == `N/A` && quotes.stocks[i].MarketCapX != `N/A` {
			quotes.stocks[i].MarketCap = quotes.stocks[i].MarketCapX
		}
		//
		// Stock is advancing if the change is not negative (i.e. $0.00
		// is also "advancing").
		//
		quotes.stocks[i].Advancing = (quotes.stocks[i].Change[0:1] != `-`)
	}

	return quotes
}

//-----------------------------------------------------------------------------
func sanitize(body []byte) []byte {
	return bytes.Replace(bytes.TrimSpace(body), []byte{'\n'}, []byte{}, -1)
}
