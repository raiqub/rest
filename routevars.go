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
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

// A RouteVars represents the route variables for specified request.
type RouteVars map[string]string

// GetObjectId tries to gets the value for specified key as int.
func (s RouteVars) GetInt(key string) (int, bool) {
	val, ok := s[key]
	if !ok {
		return 0, false
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return intVal, true
}

// GetObjectId tries to gets the value for specified key as BSON ObjectId.
func (s RouteVars) GetObjectId(key string) (bson.ObjectId, bool) {
	val, ok := s[key]
	if !ok {
		return "", false
	}

	if !bson.IsObjectIdHex(val) {
		return "", false
	}

	return bson.ObjectIdHex(val), true
}

// GetObjectId tries to gets the value for specified key as string.
func (s RouteVars) GetString(key string) (string, bool) {
	val, ok := s[key]
	if !ok {
		return "", false
	}

	return val, true
}
