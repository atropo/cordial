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

/*
Geneos configuration data model, sparsely populated

As the requirements for the various configuration items increases, just
add more to these structs

Beware the complexities of encoding/xml tags

Order of fields in structs is important, otherwise the Gateway validation
will warn of wrong ordering
*/

package geneos

import (
	"encoding/xml"
	"time"
)

type Gateway struct {
	XMLName         xml.Name `xml:"gateway"`
	Compatibility   int      `xml:"compatibility,attr"`
	XMLNs           string   `xml:"xmlns:xsi,attr"`                     // http://www.w3.org/2001/XMLSchema-instance
	XSI             string   `xml:"xsi:noNamespaceSchemaLocation,attr"` // http://schema.itrsgroup.com/GA5.12.0-220125/gateway.xsd
	ManagedEntities *ManagedEntities
	Types           *Types
	Samplers        *Samplers
	Environments    *Environments
}

type ManagedEntities struct {
	XMLName            xml.Name `xml:"managedEntities"`
	ManagedEntityGroup struct {
		XMLName    xml.Name        `xml:"managedEntityGroup"`
		Name       string          `xml:"name,attr"`
		Attributes []Attribute     `xml:",omitempty"`
		Vars       []Vars          `xml:",omitempty"`
		Entities   []ManagedEntity `xml:",omitempty"`
	}
}

type ManagedEntity struct {
	XMLName xml.Name `xml:"managedEntity"`
	Name    string   `xml:"name,attr"`
	Probe   struct {
		Name     string         `xml:"ref,attr"`
		Timezone *time.Location `xml:"-"`
	} `xml:"probe"`
	Attributes []Attribute `xml:",omitempty"`
	AddTypes   struct {
		XMLName xml.Name    `xml:"addTypes"`
		Types   []Reference `xml:"type,omitempty"`
	}
	Vars []Vars `xml:",omitempty"`
}

type Attribute struct {
	XMLName xml.Name `xml:"attribute"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

type Types struct {
	XMLName xml.Name `xml:"types"`
	Group   struct {
		XMLName xml.Name `xml:"typeGroup"`
		Name    string   `xml:"name,attr"`
		Types   []Type
	}
}

type Type struct {
	XMLName      xml.Name    `xml:"type"`
	Name         string      `xml:"name,attr"`
	Environments []Reference `xml:"environment,omitempty"`
	Vars         []Vars      `xml:",omitempty"`
	Samplers     []Reference `xml:"sampler,omitempty"`
}

type Environments struct {
	XMLName      xml.Name `xml:"environments"`
	Groups       []EnvironmentGroup
	Environments []Environment
}

type EnvironmentGroup struct {
	XMLName      xml.Name `xml:"environmentGroup"`
	Name         string   `xml:"name,attr"`
	Environments []Environment
}

type Environment struct {
	XMLName      xml.Name      `xml:"environment,omitempty"`
	Name         string        `xml:"name,attr"`
	Environments []Environment `xml:"environment,omitempty"`
	Vars         []Vars
}

type Samplers struct {
	XMLName      xml.Name `xml:"samplers"`
	SamplerGroup struct {
		Name          string `xml:"name,attr"`
		SamplerGroups []interface{}
		Samplers      []Sampler
	} `xml:"samplerGroup,omitempty"`
	Samplers []Sampler `xml:",omitempty"`
}

// A Sampler is a Geneos Sampler structure. The Plugin field should be
// populated with a pointer to a Plugin struct of the wanted type.
type Sampler struct {
	XMLName   xml.Name          `xml:"sampler"`
	Name      string            `xml:"name,attr"`
	Comment   string            `xml:",comment"`
	Group     *SingleLineString `xml:"var-group,omitempty"`
	Interval  *Value            `xml:"sampleInterval,omitempty"`
	Plugin    interface{}       `xml:"plugin"`
	Dataviews []Dataview        `xml:"dataviews>dataview,omitempty"`
}

// Gateway-SQL

type GatewaySQLPlugin struct {
	Setup  *SingleLineString `xml:"Gateway-sql>setupSql>sql"`
	Tables []GatewaySQLTable `xml:"Gateway-sql>tables>xpath"`
	Views  []GWSQLView       `xml:"Gateway-sql>views>view"`
}

type GatewaySQLTable struct {
	XMLName xml.Name          `xml:"xpath"`
	Name    *SingleLineString `xml:"tableName"`
	XPaths  []string          `xml:"xpaths>xpath"`
	Columns []GWSQLColumn     `xml:"columns>column"`
}

type GWSQLColumn struct {
	Name  *SingleLineString `xml:"name"`
	XPath string            `xml:"xpath"`
	Type  string            `xml:"type"`
}

type GWSQLView struct {
	XMLName  xml.Name          `xml:"view"`
	ViewName *SingleLineString `xml:"name"`
	SQL      *SingleLineString `xml:"sql"`
}

// FTM

type FTMPlugin struct {
	Files                      []FTMFile `xml:"ftm>files>file"`
	ConsistentDateStamps       *Value    `xml:"ftm>consistentDateStamps,omitempty"`
	DisplayTimeInISO8601Format *Value    `xml:"ftm>displayTimeInIso8601Format,omitempty"`
	ShowActualFilename         *Value    `xml:"ftm>showActualFilename,omitempty"`
	DelayUnit                  string    `xml:"ftm>delayUnit"`
	SizeUnit                   string    `xml:"ftm>sizeUnit"`
}

type FTMFile struct {
	XMLName         xml.Name            `xml:"file"`
	Path            *SingleLineString   `xml:"path"`
	AdditionalPaths *FTMAdditionalPaths `xml:"additionalPaths,omitempty"`
	ExpectedArrival *Value              `xml:"expectedArrival,omitempty"`
	ExpectedPeriod  *struct {
		Period string `xml:",innerxml"`
	} `xml:"expectedPeriod,omitempty"`
	TZOffset         *Value            `xml:"tzOffset,omitempty"`
	MonitoringPeriod interface{}       `xml:"monitoringPeriod"`
	Alias            *SingleLineString `xml:"alias"`
}

type MonitoringPeriodAlias struct {
	Alias string `xml:"periodAlias"`
}

type MonitoringPeriodStart struct {
	PeriodStart *Value `xml:"periodStart,omitempty"`
}

type FTMAdditionalPaths struct {
	Paths []*SingleLineString `xml:"additionalPath"`
}

// FKM

type FKMPlugin struct {
	Display *FKMDisplay `xml:"display,omitempty"`
	Files   []FKMFile
}

type FKMDisplay struct {
	TriggerMode string `xml:"triggerMode,omitempty"`
}

type FKMFile struct {
	Filename             *Value `xml:"source>filename,omitempty"`
	Stream               *Value `xml:"source>stream,omitempty"`
	Tables               []FKMTable
	ClearTime            *Value `xml:"clearTime"`
	DefaultKeyClkearTime *Value `xml:"defaultKeyClearTime"`
	Rewind               *Value `xml:"rewind"`
	Alias                *Value `xml:"alias"`
}

type FKMTable struct {
	XMLName  xml.Name `xml:"table"`
	Severity string   `xml:"severity"`
	KeyTable FKMKeys
}

type FKMKeys struct {
	Keys []interface{}
}

type FKMIgnoreKey struct {
	XMLName xml.Name `xml:"ignoreKey"`
	Match   FKMMatch `xml:"match"`
	// ActiveTime
}

type FKMKey struct {
	XMLName  xml.Name `xml:"key"`
	SetKey   FKMMatch
	ClearKey FKMMatch
	Message  *Value
	Severity string
}

type FKMMatch struct {
	SearchString *Value `xml:"searchString"`
	Rules        string `xml:"rules,omitempty"`
}

// SQL Toolkit

type SQLToolkitPlugin struct {
	Queries    []Query      `xml:"sql-toolkit>queries>query"`
	Connection DBConnection `xml:"sql-toolkit>connection"`
}

type Query struct {
	Name *SingleLineString `xml:"name"`
	SQL  *SingleLineString `xml:"sql"`
}

type DBConnection struct {
	MySQL                     *MySQL            `xml:"database>mysql,omitempty"`
	SQLServer                 *SQLServer        `xml:"database>sqlServer,omitempty"`
	Sybase                    *Sybase           `xml:"database>sybase,omitempty"`
	Username                  *SingleLineString `xml:"var-userName"`
	Password                  *SingleLineString `xml:"password"`
	CloseConnectionAfterQuery *Value            `xml:"closeConnectionAfterQuery,omitempty"`
}

type MySQL struct {
	ServerName *SingleLineString `xml:"var-serverName"`
	DBName     *SingleLineString `xml:"var-databaseName"`
	Port       *SingleLineString `xml:"var-port"`
}

type SQLServer struct {
	ServerName *SingleLineString `xml:"var-serverName"`
	DBName     *SingleLineString `xml:"var-databaseName"`
	Port       *SingleLineString `xml:"var-port"`
}

type Sybase struct {
	InstanceName *SingleLineString `xml:"var-instanceName"`
	DBName       *SingleLineString `xml:"var-databaseName"`
}

type ToolkitPlugin struct {
	SamplerScript        *SingleLineString     `xml:"toolkit>samplerScript"`
	EnvironmentVariables []EnvironmentVariable `xml:"toolkit>environmentVariables>variable"`
}

type EnvironmentVariable struct {
	Name  string            `xml:"name"`
	Value *SingleLineString `xml:"value"`
}

// API Plugin

type APIPlugin struct {
	Parameters  []Parameter       `xml:"api>parameters>parameter"`
	SummaryView *SingleLineString `xml:"api>showSummaryView>always>viewName,omitempty"`
}

type Parameter struct {
	Name  string            `xml:"name"`
	Value *SingleLineString `xml:"value"`
}

// API Streams Plugin

type APIStreamsPlugin struct {
	Streams    *Streams `xml:"api-streams>streams"`
	CreateView *Value   `xml:"api-streams>createView,omitempty"`
}

type Streams struct {
	XMLName xml.Name            `xml:"streams"`
	Stream  []*SingleLineString `xml:"stream"`
}

type Dataview struct {
	Name      string `xml:"name,attr"`
	Additions DataviewAdditions
}

type DataviewAdditions struct {
	XMLName   xml.Name `xml:"additions"`
	Headlines *Value   `xml:"var-headlines,omitempty"`
	Columns   *Value   `xml:"var-columns,omitempty"`
	Rows      *Value   `xml:"var-rows,omitempty"`
}

type DataviewAddition struct {
	XMLName xml.Name          `xml:"data"`
	Name    *SingleLineString `xml:"headline,omitempty"`
}

// Rules

type Rules struct {
	XMLName    xml.Name `xml:"rules"`
	RuleGroups []interface{}
	Rules      []interface{}
}

type RuleGroups struct {
	XMLName xml.Name `xml:"ruleGroup"`
	Name    string   `xml:"name,attr"`
}

type Rule struct {
	XMLName      xml.Name `xml:"rule"`
	Name         string   `xml:"name,attr"`
	Targets      []string `xml:"targets>target"`
	Priority     int      `xml:"priority"`
	Ifs          []interface{}
	Transactions []interface{}
}
