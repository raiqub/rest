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
	"fmt"

	"github.com/gin-gonic/gin"
	"gopkg.in/raiqub/web.v0"
)

// RecoverHandlerJson is a HTTP request middleware that captures panic errors
// and returns it as HTTP JSON response.
func RecoverHandlerJson(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			jerr := web.NewJSONError().
				FromError(fmt.Errorf("panic: %+v", err)).
				Build()
			c.JSON(jerr.Status, jerr)
		}
	}()

	c.Next()
}
