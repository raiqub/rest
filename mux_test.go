/*
 * Copyright (C) 2015 Fabrício Godoy <skarllot@gmail.com>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rqhttp "github.com/raiqub/http"
)

const (
	TRIM_CUT_SET = "\n\r "
)

func TestResourceRouting(t *testing.T) {
	resources := []*resource{
		&resource{
			"a",
			map[string]string{
				"0": "lorem",
				"1": "ipsum",
				"2": "dolor",
				"3": "sit",
				"4": "amet",
				"5": "consectetur",
			},
		},
		&resource{
			"b",
			map[string]string{
				"0": "adipiscing",
				"1": "elit.",
				"2": "Sed",
				"3": "tortor",
				"4": "justo",
				"5": "dui",
			},
		},
	}
	router := NewRest()
	for _, r := range resources {
		router.AddResource(r)
	}

	ts := httptest.NewServer(router)
	defer ts.Close()

	for _, r := range resources {
		aJson, _ := json.Marshal(r.data)
		client := http.Client{}
		ret, err := client.Get(fmt.Sprintf("%s/%s", ts.URL, r.name))
		if err != nil {
			t.Fatal("Error getting resource list:", err.Error())
		}

		body, err := ioutil.ReadAll(ret.Body)
		if err != nil {
			t.Fatal("Error reading body contents:", err.Error())
		}
		ret.Body.Close()

		body = bytes.Trim(body, TRIM_CUT_SET)
		if bytes.Compare(body, aJson) != 0 {
			t.Logf("Expected: %s", aJson)
			t.Logf("Got: %s", body)
			t.Errorf("Unexpected output when GETting /%s", r.name)
		}

		for key, value := range r.data {
			aJson, _ = json.Marshal(value)
			ret, err = client.Get(fmt.Sprintf("%s/%s/%s", ts.URL, r.name, key))
			if err != nil {
				t.Fatalf("Error getting resource item '%s': %v", key, err)
			}

			body, err = ioutil.ReadAll(ret.Body)
			if err != nil {
				t.Fatal("Error reading body contents:", err.Error())
			}
			ret.Body.Close()

			body = bytes.Trim(body, TRIM_CUT_SET)
			if bytes.Compare(body, aJson) != 0 {
				t.Logf("Expected: %s", value)
				t.Logf("Got: %s", body)
				t.Errorf("Unexpected output when GETting /%s/%s", r.name, key)
			}
		}
	}
}

type resource struct {
	name string
	data map[string]string
}

func newResource(name string) *resource {
	return &resource{
		name,
		make(map[string]string),
	}
}

func (self *resource) GetItem(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := Vars(r)["id"]

	ret, ok := self.data[id]
	if !ok {
		jerr := rqhttp.NewJsonErrorFromError(http.StatusNotFound,
			fmt.Errorf("Id '%s' not Found", id))
		rqhttp.JsonWrite(w, jerr.Status, jerr)
		return
	}

	rqhttp.JsonWrite(w, http.StatusOK, ret)
}

func (self *resource) GetList(
	w http.ResponseWriter,
	r *http.Request,
) {
	rqhttp.JsonWrite(w, http.StatusOK, self.data)
}

func (self *resource) Routes() Routes {
	return Routes{
		Route{
			"GetList",
			"GET",
			fmt.Sprintf("/%s", self.name),
			false,
			self.GetList,
		},
		Route{
			"GetItem",
			"GET",
			fmt.Sprintf("/%s/{id}", self.name),
			false,
			self.GetItem,
		},
	}
}
