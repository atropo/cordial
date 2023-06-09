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

package geneos

import "encoding/xml"

// These types represent those found in a netprobe setup file (SAN or
// floating) and not a probe config in the gateway

type Netprobe struct {
	XMLName          xml.Name          `xml:"netprobe"`
	Compatibility    int               `xml:"compatibility,attr"`                 // 1
	XMLNs            string            `xml:"xmlns:xsi,attr"`                     // http://www.w3.org/2001/XMLSchema-instance
	XSI              string            `xml:"xsi:noNamespaceSchemaLocation,attr"` // http://schema.itrsgroup.com/GA5.12.0-220125/netprobe.xsd
	FloatingNetprobe *FloatingNetprobe `xml:"floatingProbe,omitempty"`
	PluginWhiteList  []string          `xml:"pluginWhiteList,omitempty"`
	CommandWhiteList []string          `xml:"commandWhiteList,omitempty"`
	SelfAnnounce     *SelfAnnounce     `xml:"selfAnnounce,omitempty"`
}

type FloatingNetprobe struct {
	Enabled                  bool       `xml:"enabled"`
	RetryInterval            int        `xml:"retryInterval,omitempty"`
	RequireReverseConnection bool       `xml:"requireReverseConnection,omitempty"`
	ProbeName                string     `xml:"probeName"`
	Gateways                 []Gateways `xml:"gateways"`
}

type Gateways struct {
	XMLName  xml.Name `xml:"gateway"`
	Hostname string   `xml:"hostname"`
	Port     int      `xml:"port,omitempty"`
	Secure   bool     `xml:"secure,omitempty"`
}

type SelfAnnounce struct {
	Enabled                  bool              `xml:"enabled"`
	RetryInterval            int               `xml:"retryInterval,omitempty"`
	RequireReverseConnection bool              `xml:"requireReverseConnection,omitempty"`
	ProbeName                string            `xml:"probeName"`
	EncodedPassword          string            `xml:"encodedPassword,omitempty"`
	RESTAPIHTTPPort          int               `xml:"restApiHttpPort,omitempty"`
	RESTAPIHTTPSPort         int               `xml:"restApiHttpsPort,omitempty"`
	CyberArkApplicationID    string            `xml:"cyberArkApplicationID,omitempty"`
	CyberArkSDKPath          string            `xml:"cyberArkSdkPath,omitempty"`
	ManagedEntity            *SAManagedEntity  `xml:"managedEntity,omitempty"`
	ManagedEntities          []SAManagedEntity `xml:"managedEntities,omitempty"`
	CollectionAgent          *CollectionAgent  `xml:"collectionAgent,omitempty"`
	DynamicEntities          *DynamicEntities  `xml:"dynamicEntities,omitempty"`
	Gateways                 []Gateways        `xml:"gateways"`
}

type SAManagedEntity struct {
	XMLName    xml.Name    `xml:"managedEntity"`
	Name       string      `xml:"name"`
	Attributes []Attribute `xml:"attributes,omitempty"`
	Vars       []Vars      `xml:"variables,omitempty"`
	Types      []PlainType `xml:"types,omitempty"`
}

type PlainType struct {
	Type string `xml:"type"`
}

type CollectionAgent struct {
	Start        bool   `xml:"start,omitempty"`
	JVMArgs      string `xml:"jvmArgs,omitempty"`
	HealthPort   int    `xml:"healthPort,omitempty"`
	ReporterPort int    `xml:"reporterPort,omitempty"`
	Detached     bool   `xml:"detached"`
}

type DynamicEntities struct {
	MappingType []string `xml:"mappingType,omitempty"`
}
