// Copyright (C) 2024  Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package findat implements financial data fetching.
package findat

import (
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/quote"
)

type Client struct {
	cache map[string]*finance.Quote
}

func NewClient() *Client {
	return &Client{
		cache: make(map[string]*finance.Quote),
	}
}

func (c *Client) GetQuote(symbol string) (*finance.Quote, error) {
	q, ok := c.cache[symbol]
	if ok {
		return q, nil
	}
	q, err := quote.Get(symbol)
	if err != nil {
		return nil, err
	}
	c.cache[symbol] = q
	return q, nil
}
