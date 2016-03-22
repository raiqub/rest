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
	"sync"

	"github.com/gorilla/mux"
	"gopkg.in/raiqub/web.v0"
)

// A Rest register resources and middlewares for a HTTP handler.
//
// It implements the http.Handler interface, so it can be registered to serve
// requests.
type Rest struct {
	router        *mux.Router
	routes        Routes
	middlePublic  web.Chain
	middlePrivate web.Chain
	cors          *CORSHandler
	prepare       sync.Once
}

// NewRest returns a new instance of Rest router.
func NewRest() *Rest {
	return &Rest{
		nil,
		make(Routes, 0),
		make(web.Chain, 0),
		make(web.Chain, 0),
		nil,
		sync.Once{},
	}
}

// AddMiddlewarePrivate adds a layer to handle private resource requests.
func (rest *Rest) AddMiddlewarePrivate(m web.MiddlewareFunc) {
	rest.middlePrivate = append(rest.middlePrivate, m)
}

// AddMiddlewarePublic adds a layer to handle public resource requests.
func (rest *Rest) AddMiddlewarePublic(m web.MiddlewareFunc) {
	rest.middlePublic = append(rest.middlePublic, m)
}

// AddResource adds a new REST-resource to handle requests.
func (rest *Rest) AddResource(r Routable) {
	rest.routes = append(rest.routes, r.Routes()...)
}

// EnableCORS allows to current API support CORS specification.
func (rest *Rest) EnableCORS() {
	rest.cors = NewCORSHandler()
}

func (rest *Rest) initRouter() {
	if rest.cors != nil {
		rest.routes = append(rest.routes,
			rest.cors.CreatePreflight(rest.routes)...)
	}

	rest.router = mux.NewRouter().StrictSlash(true)
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
}

// Routes returns the routes from registered resources.
func (rest *Rest) ResourcesRoutes() Routes {
	result := make(Routes, len(rest.routes))
	copy(result, rest.routes)

	return result
}

// ServeHTTP dispatches the handler registered in the matched route.
func (rest *Rest) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rest.prepare.Do(rest.initRouter)
	rest.router.ServeHTTP(w, req)
}

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) RouteVars {
	return mux.Vars(r)
}
