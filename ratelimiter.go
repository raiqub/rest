/*
 * Copyright 2016 Fabrício Godoy
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rest

import "gopkg.in/raiqub/data.v0"

// A RateLimiter limits the number of requests per timing window.
type RateLimiter struct {
	stack     *RateLimiter
	store     data.Store
	threshold int
}

// NewRateLimiter creates a new instance of RateLimiter based on specified store
// and number of requests per timing window.
func NewRateLimiter(store data.Store, threshold int) *RateLimiter {
	return &RateLimiter{
		store:     store,
		threshold: threshold,
	}
}

// AddCall register a new request for specified identifier and returns how many
// levels the identifier overflow the threshold.
func (lmt *RateLimiter) AddCall(id string) int {
	var level int
	count, err := lmt.store.Increment(id)
	if err != nil {
		return -1
	}

	if count >= lmt.threshold {
		level++

		if lmt.stack != nil {
			level += lmt.stack.AddCall(id)
		}
	}

	return level
}

// Stack sets a rate limiter for the next level.
func (lmt *RateLimiter) Stack(another *RateLimiter) {
	lmt.stack = another
}
