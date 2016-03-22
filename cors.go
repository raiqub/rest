/*
 * Copyright 2015 Fabrício Godoy
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
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/raiqub/dot.v1"
	"gopkg.in/raiqub/web.v0"
)

const (
	DEFAULT_CORS_PREFLIGHT_METHOD = "OPTIONS"
	DEFAULT_CORS_MAX_AGE          = time.Hour * 24 / time.Second
	DEFAULT_CORS_METHODS          = "OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, CONNECT"
	DEFAULT_CORS_ORIGIN           = "*"
)

// A CORSHandler allows to create a CORS-able API.
type CORSHandler struct {
	PredicateOrigin dot.PredicateStringFunc
	Headers         []string
	ExposedHeaders  []string
}

// NewCORSHandler creates a new CORSHandler with default values.
func NewCORSHandler() *CORSHandler {
	return &CORSHandler{
		PredicateOrigin: dot.TrueForAll,
		Headers: []string{
			"Origin", "X-Requested-With", "Content-Type",
			"Accept", "Authorization",
		},
		ExposedHeaders: make([]string, 0),
	}
}

// CreatePreflight creates HTTP routes that handles pre-flight requests.
func (s *CORSHandler) CreatePreflight(routes Routes) Routes {
	list := make(Routes, 0, len(routes))
	hList := make(map[string]*CORSPreflight, len(routes))
	for _, v := range routes {
		preflight, ok := hList[v.Path]
		if !ok {
			preflight = &CORSPreflight{
				*s,
				make([]string, 0, 1),
				v.MustAuth,
			}
			hList[v.Path] = preflight
		}

		preflight.Methods = append(preflight.Methods, v.Method)
		if v.MustAuth {
			preflight.UseCredentials = true
		}
	}

	for k, v := range hList {
		list = append(list, Route{
			Name:       "",
			Method:     DEFAULT_CORS_PREFLIGHT_METHOD,
			Path:       k,
			MustAuth:   v.UseCredentials,
			ActionFunc: v.ServeHTTP,
		})
	}
	return list
}

// A CORSPreflight represents a HTTP server that handles pre-flight requests.
type CORSPreflight struct {
	CORSHandler
	Methods        []string
	UseCredentials bool
}

// ServeHTTP handle a pre-flight request.
func (s *CORSPreflight) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := web.NewHeader().Origin().Read(r.Header)
	status := http.StatusBadRequest
	msg := ""
	defer func() {
		w.WriteHeader(status)
		w.Write([]byte(msg))
	}()

	if origin.Value != "" {
		if !s.PredicateOrigin(origin.Value) {
			status = http.StatusForbidden
			return
		}

		method := web.NewHeader().
			AccessControlRequestMethod().
			Read(r.Header).
			Value
		header := strings.Split(
			web.NewHeader().
				AccessControlRequestHeaders().
				Read(r.Header).
				Value,
			", ",
		)
		if len(header) == 1 && header[0] == "" {
			header = []string{}
		}

		if !dot.StringSlice(s.Methods).Exists(method, false) {
			msg = "Method not allowed"
			return
		}

		if len(s.Headers) == 0 {
			web.NewHeader().
				AccessControlAllowHeaders().
				Write(w.Header())
		} else {
			if len(header) > 0 &&
				!dot.StringSlice(s.Headers).ExistsAll(header, true) {
				msg = "Header not allowed"
				return
			}
			web.NewHeader().
				AccessControlAllowHeaders().
				SetValue(strings.Join(s.Headers, ", ")).
				Write(w.Header())
		}

		web.NewHeader().
			AccessControlAllowMethods().
			SetValue(strings.Join(s.Methods, ", ")).
			Write(w.Header())
		web.NewHeader().
			AccessControlAllowOrigin().
			SetValue(origin.Value).
			Write(w.Header())
		web.NewHeader().
			AccessControlAllowCredentials().
			SetValue(strconv.FormatBool(s.UseCredentials)).
			Write(w.Header())
		// Optional
		web.NewHeader().
			AccessControlMaxAge().
			SetValue(strconv.Itoa(int(DEFAULT_CORS_MAX_AGE))).
			Write(w.Header())
		status = http.StatusOK
	} else {
		status = http.StatusNotFound
	}
}

// A CORSMiddleware represents a HTTP middleware that handle HTTP headers for
// CORS-able API.
type CORSMiddleware struct {
	CORSHandler
	UseCredentials bool
}

// Handle is a HTTP handler for CORS-able API.
func (s *CORSMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		origin := web.NewHeader().Origin().Read(r.Header)
		if r.Method != DEFAULT_CORS_PREFLIGHT_METHOD && origin.Value != "" {
			if !s.PredicateOrigin(origin.Value) {
				return
			}

			web.NewHeader().
				AccessControlAllowOrigin().
				SetValue(origin.Value).
				Write(w.Header())
			web.NewHeader().
				AccessControlAllowCredentials().
				SetValue(strconv.FormatBool(s.UseCredentials)).
				Write(w.Header())
			if len(s.Headers) > 0 {
				web.NewHeader().
					AccessControlAllowHeaders().
					SetValue(strings.Join(s.Headers, ", ")).
					Write(w.Header())
			} else {
				web.NewHeader().
					AccessControlAllowHeaders().
					Write(w.Header())
			}
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
