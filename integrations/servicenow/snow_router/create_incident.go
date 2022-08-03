/*
Copyright © 2022 ITRS Group

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/itrs-group/cordial/integrations/servicenow/snow"
)

func CreateIncident(sys_id string, incident Incident) (incident_number string, err error) {
	// var ok bool
	var postbytes []byte
	var result snow.ResultDetail
	// var stateid string

	// Initialize ServiceNow Connection
	s := InitializeConnection()

	incident["cmdb_ci"] = sys_id

	// this has to bypass default settings in caller
	if incident["text"] != "" {
		incident["description"] = incident["text"]
		delete(incident, "text")
	}

	postbytes, err = json.Marshal(incident)
	if err != nil {
		fmt.Println(err)
	}
	result, err = s.POST(postbytes, "", "number", "", "", "").QueryTableSingle(cf.ServiceNow.IncidentTable)
	if err != nil {
		return
	} else {
		incident_number = result["number"]
	}

	return
}
