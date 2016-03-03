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

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	CloudflareRanges = []string{
		"https://www.cloudflare.com/ips-v4",
		"https://www.cloudflare.com/ips-v6",
	}
)

type Whitelist struct {
	cidrs    []*net.IPNet
	listUrls []string
	interval time.Duration
	mutex    sync.RWMutex
}

func NewWhitelist(interval time.Duration, urls []string) *Whitelist {
	w := &Whitelist{
		make([]*net.IPNet, 0),
		urls,
		interval,
		sync.RWMutex{},
	}

	w.fetch()
	if len(w.listUrls) == 0 {
		return nil
	}

	if interval > 0 {
		go w.continuousFetch()
	}

	return w
}

func (w *Whitelist) fetch() {
	list := make([]*net.IPNet, 0)
	for _, url := range w.listUrls {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			ip := strings.TrimSpace(scanner.Text())
			if len(ip) == 0 {
				continue
			}

			_, cidr, err := net.ParseCIDR(ip)
			if err != nil {
				continue
			}

			list = append(list, cidr)
		}
		resp.Body.Close()
	}

	if len(list) > 0 {
		w.mutex.Lock()
		w.cidrs = list
		w.mutex.Unlock()
	}
}

func (w *Whitelist) Handler(c *gin.Context) {
	if w.interval > 0 {
		w.mutex.RLock()
		defer w.mutex.RUnlock()
	}

	if len(w.cidrs) == 0 {
		return
	}

	remoteip := remoteIP(c)
	if len(remoteip) == 0 {
		return
	}

	netip := net.ParseIP(remoteip)
	if netip == nil {
		return
	}

	for _, cidr := range w.cidrs {
		if cidr.Contains(netip) {
			return
		}
	}

	c.String(http.StatusInternalServerError, "Bad Host")
	c.Abort()
}

func (w *Whitelist) continuousFetch() {
	for {
		<-time.After(w.interval)
		w.fetch()
	}
}

func remoteIP(c *gin.Context) string {
	if ip, _, err := net.SplitHostPort(
		strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
