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

	"github.com/gorilla/mux"
	rqhttp "github.com/raiqub/http"
)

type Rest struct {
	router        *mux.Router
	routes        Routes
	detached      Routes
	middlePublic  rqhttp.Chain
	middlePrivate rqhttp.Chain
	cors          *CORSHandler
}

func NewRest() *Rest {
	return &Rest{
		nil,
		make(Routes, 0),
		make(Routes, 0),
		make(rqhttp.Chain, 0),
		make(rqhttp.Chain, 0),
		nil,
	}
}

func (rest *Rest) AddMiddlewarePrivate(m rqhttp.HttpMiddlewareFunc) {
	rest.middlePrivate = append(rest.middlePrivate, m)
}

func (rest *Rest) AddMiddlewarePublic(m rqhttp.HttpMiddlewareFunc) {
	rest.middlePublic = append(rest.middlePublic, m)
}

func (rest *Rest) AddResource(r Routable) {
	rest.routes = append(rest.routes, r.Routes()...)
}

func (rest *Rest) AddResourceDetached(r Routable) {
	rest.detached = append(rest.detached, r.Routes()...)
}

func (rest *Rest) EnableCORS() {
	rest.cors = NewCORSHandler()
}

func (rest *Rest) ListenAndServe(addr string) error {
	if rest.cors != nil {
		rest.routes = append(rest.routes,
			rest.cors.CreatePreflight(rest.routes)...)
	}

	rest.router = mux.NewRouter()
	for _, r := range rest.routes {
		middlewares := rest.middlePublic
		if r.MustAuth == true {
			middlewares = rest.middlePrivate
		}

		if rest.cors != nil {
			middlewares = append(middlewares,
				(&CORSMiddleware{*rest.cors, r.MustAuth}).Handle)
		}

		rest.router.
			Methods(r.Method).
			Path(r.Path).
			Name(r.Name).
			Handler(middlewares.Get(r.ActionFunc))
	}

	if len(rest.detached) > 0 {
		for _, r := range rest.detached {
			rest.router.
				Methods(r.Method).
				Path(r.Path).
				Name(r.Name).
				Handler(r.ActionFunc)
		}
	}

	return http.ListenAndServe(addr, rest.router)
}

func (rest *Rest) ResourcesRoutes() Routes {
	result := make(Routes, len(rest.routes))
	copy(result, rest.routes)
	
	return result
}
