//
// Copyright © 2016 Bryan T. Meyers <bmeyers@datadrake.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package providers

import (
	"encoding/json"
	"github.com/DataDrake/cuppa/results"
	"github.com/DataDrake/cuppa/utility"
	"net/http"
	"net/url"
	"time"
)

var cpanDistAPI = "http://search.cpan.org/api/dist"
var cpanSrcRoot = "http://search.cpan.org/CPAN/authors/id"

type cpanRelease struct {
	Dist     string `json:"dist"`
	Archive  string `json:"archive"`
	Cpanid   string `json:"cpanid"`
	Version  string `json:"version"`
	Released string `json:"released"`
}

func (cr *cpanRelease) Convert() *results.Result {
	r := &results.Result{}
	r.Name = cr.Dist
	r.Version = cr.Version
	r.Published, _ = time.Parse(time.RFC3339, cr.Released)
	r.Location, _ = url.Parse(utility.URLJoin(cpanSrcRoot, cr.Cpanid[0:1], cr.Cpanid[0:2], cr.Cpanid, cr.Archive))
	return r
}

type cpanResultSet struct {
	Releases []cpanRelease `json:"releases"`
}

func (crs *cpanResultSet) Convert(name string) *results.ResultSet {
	rs := results.NewResultSet(name)
	for _, rel := range crs.Releases {
		r := rel.Convert()
		rs.AddResult(r)
	}
	return rs
}

/*
CPANProvider is the upstream provider interface for CPAN
*/
type CPANProvider struct{}

/*
Search finds all matching releases for a CPAN package
*/
func (c CPANProvider) Search(Name string) (rs *results.ResultSet, s results.Status) {
	//Query the API
	resp, err := http.Get(utility.URLJoin(cpanDistAPI, Name))
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	//Translate Status Code
	switch resp.StatusCode {
	case 200:
		s = results.OK
	case 404:
		s = results.NotFound
	default:
		s = results.Unavailable
	}

	//Fail if not OK
	if s != results.OK {
		return
	}

	dec := json.NewDecoder(resp.Body)
	crs := &cpanResultSet{}
	err = dec.Decode(crs)
	if err != nil {
		panic(err.Error())
	}
	rs = crs.Convert(Name)
	return
}
