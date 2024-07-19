package models

import (
	"encoding/xml"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScrapeGroup represents a group of scrapes with associated fields and endpoints
type ScrapeGroup struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Fields        []Field            `json:"fields" bson:"fields"`
	Endpoints     []Endpoint         `json:"endpoints" bson:"endpoints"`
	WithThumbnail bool               `json:"withThumbnail" bson:"withThumbnail"`
	VersionTag    string             `json:"versionTag" bson:"versionTag"`
	Created       primitive.DateTime `json:"created" bson:"created"`
	Updated       primitive.DateTime `json:"updated" bson:"updated"`
}

type ArchivedScrapeGroup struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OriginalID    primitive.ObjectID `json:"originalId" bson:"originalId"`
	Name          string             `json:"name" bson:"name"`
	Fields        []Field            `json:"fields" bson:"fields"`
	Endpoints     []Endpoint         `json:"endpoints" bson:"endpoints"`
	WithThumbnail bool               `json:"withThumbnail" bson:"withThumbnail"`
	VersionTag    string             `json:"versionTag" bson:"versionTag"`
	Created       primitive.DateTime `json:"created" bson:"created"`
	Updated       primitive.DateTime `json:"updated" bson:"updated"`
}

func (sg ScrapeGroup) GetEndpointById(id string) *Endpoint {
	for _, endpoint := range sg.Endpoints {
		if endpoint.ID == id {
			return &endpoint
		}
	}
	return nil
}

func (sc ScrapeGroup) GetFieldById(id string) *Field {
	for _, field := range sc.Fields {
		if field.ID == id {
			return &field
		}
	}
	return nil

}

func (sg *ScrapeGroup) DeleteEndpoint(id string) {
	var newEndpoints []Endpoint
	for _, endpoint := range sg.Endpoints {
		if endpoint.ID != id {
			newEndpoints = append(newEndpoints, endpoint)
		}
	}
	sg.Endpoints = newEndpoints
}

type ScrapeGroupLocal struct {
	ID        string     `json:"id" bson:"id"`
	Name      string     `json:"name" bson:"name"`
	Fields    []Field    `json:"fields" bson:"fields"`
	Endpoints []Endpoint `json:"endpoints" bson:"endpoints"`
}

// Field represents a single field with its type and name
type Field struct {
	ID              string    `json:"id" bson:"id"`
	Name            string    `json:"name" bson:"name"`
	Key             string    `json:"key" bson:"key"`
	Type            FieldType `json:"type" bson:"type"`
	IsFullyEditable bool      `json:"isFullyEditable" bson:"isFullyEditable"`
	Order           int       `json:"order" bson:"order"`
}

// FieldType defines the type of a field
type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeImage  FieldType = "image"
	FieldTypeLink   FieldType = "link"
	FieldTypeNumber FieldType = "number"
)

// Endpoint represents an endpoint with its configuration
type Endpoint struct {
	ID                   string           `json:"id" bson:"id"`
	Name                 string           `json:"name" bson:"name"`
	URL                  string           `json:"url" bson:"url"`
	PaginationConfig     PaginationConfig `json:"paginationConfig" bson:"paginationConfig"`
	MainElementSelector  string           `json:"mainElementSelector" bson:"mainElementSelector"`
	DetailFieldSelectors []FieldSelector  `json:"detailFieldSelectors" bson:"detailFieldSelectors"`
	Interval             string           `json:"interval,omitempty" bson:"interval,omitempty"`
	Active               bool             `json:"active,omitempty" bson:"active,omitempty"`
	LastScraped          time.Time        `json:"lastScraped,omitempty" bson:"lastScraped,omitempty"`
	Status               ScrapeStatus     `json:"status,omitempty" bson:"status,omitempty"`
}

type PaginationConfigType string

const (
	PaginationConfigTypeNone         PaginationConfigType = "none"
	PaginationConfigTypeUrlParameter PaginationConfigType = "url_parameter"
	PaginationConfigTypePath         PaginationConfigType = "url_path"
)

// PaginationConfig represents pagination configuration for an endpoint
type PaginationConfig struct {
	Type             string  `json:"type" bson:"type"`
	Parameter        string  `json:"parameter,omitempty" bson:"parameter,omitempty"`
	Start            int     `json:"start,omitempty" bson:"start,omitempty"`
	End              int     `json:"end,omitempty" bson:"end,omitempty"`
	Step             int     `json:"step,omitempty" bson:"step,omitempty"`
	UrlRegexToInsert *string `json:"urlRegexToInsert,omitempty" bson:"urlRegexToInsert,omitempty"`
}

// ScrapeStatus defines the status of a scrape
type ScrapeStatus string

const (
	ScrapeStatusIdle    ScrapeStatus = "idle"
	ScrapeStatusRunning ScrapeStatus = "running"
)

type SelectorStatusValue string

const (
	SelectorStatusOk          SelectorStatusValue = "ok"
	SelectorStatusNeedsUpdate SelectorStatusValue = "needs_update"
	SelectorStatusNew         SelectorStatusValue = "new"
)

// FieldSelector represents a selector for a field
type FieldSelector struct {
	ID                   string              `json:"id" bson:"id"`
	FieldID              string              `json:"fieldId" bson:"fieldId"`
	Selector             string              `json:"selector" bson:"selector"`
	Regex                string              `json:"regex" bson:"regex"`
	AttributeToGet       string              `json:"attributeToGet" bson:"attributeToGet"`
	RegexMatchIndexToUse int                 `json:"regexMatchIndexToUse" bson:"regexMatchIndexToUse"`
	SelectorStatus       SelectorStatusValue `json:"selectorStatus" bson:"selectorStatus"`
}

// SearchConfig represents search configuration for an endpoint
type SearchConfig struct {
	ID    string `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Param string `json:"param" bson:"param"`
	Value string `json:"value" bson:"value"`
}

type FieldSelectorsResponse struct {
	Field                string `json:"field" bson:"field"`
	Selector             string `json:"selector" bson:"selector"`
	Regex                string `json:"regex" bson:"regex"`
	RegexMatchIndexToUse int    `json:"regexMatchIndexToUse" bson:"regexMatchIndexToUse"`
	AttributeToGet       string `json:"attributeToGet" bson:"attributeToGet"`
}

type FieldToExtractSelectorsFor struct {
	Name   string `json:"name"`
	Key    string `json:"key"`
	Type   string `json:"type"`
	Remark string `json:"remark"`
}

type FieldSelectorsRequest struct {
	URL                         string                       `json:"url" bson:"url"`
	MainElementSelector         string                       `json:"mainElementSelector" bson:"mainElementSelector"`
	FieldsToExtractSelectorsFor []FieldToExtractSelectorsFor `json:"fieldsToExtractSelectorsFor" bson:"fieldsToExtractSelectorsFor"`
}

type ScrapeResult struct {
	ID                  primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UniqueHash          string               `json:"uniqueHash" bson:"uniqueHash"`
	EndpointID          string               `json:"endpointId" bson:"endpointId"`
	GroupId             primitive.ObjectID   `json:"groupId" bson:"groupId"`
	Fields              []ScrapeResultDetail `json:"fields" bson:"fields"`
	TimestampInitial    string               `json:"timestampInitial" bson:"timestampInitial"`
	TimestampLastUpdate string               `json:"timestampLastUpdate" bson:"timestampLastUpdate"`
	GroupVersionTag     string               `json:"groupVersionTag" bson:"groupVersionTag"`
}

type ScrapeResultDetail struct {
	ID      string      `json:"id" bson:"id"`
	FieldID string      `json:"fieldId" bson:"fieldId"`
	Value   interface{} `json:"value" bson:"value"`
}

type ScrapeResultTest struct {
	ID              primitive.ObjectID       `json:"id"`
	UniqueHash      string                   `json:"uniqueHash"`
	EndpointID      string                   `json:"endpointId"`
	GroupId         primitive.ObjectID       `json:"groupId"`
	Fields          []ScrapeResultDetailTest `json:"fields"`
	Timestamp       string                   `json:"timestamp"`
	GroupVersionTag string                   `json:"groupVersionTag"`
}

type ExportType string

const (
	ExportTypePDF   ExportType = "pdf"
	ExportTypeEXCEL ExportType = "excel"
	ExportTypeCSV   ExportType = "csv"
	ExportTypeXML   ExportType = "xml"
	ExportTypeJSON  ExportType = "json"
)

type ExportScrapeResultDetail struct {
	XMLName   xml.Name    `xml:"field"`
	ID        string      `json:"id" xml:"id"`
	FieldName string      `json:"fieldName" xml:"fieldName"`
	FieldID   string      `json:"fieldId" xml:"fieldId"`
	Value     interface{} `json:"value" xml:"value"`
}

type ExportScrapeResult struct {
	XMLName             xml.Name                   `xml:"ScrapeResult"`
	ID                  string                     `json:"id" xml:"id"`
	EndpointName        string                     `json:"endpointName" xml:"endpointName"`
	EndpointID          string                     `json:"endpointId" xml:"endpointId"`
	GroupName           string                     `json:"groupName" xml:"groupName"`
	GroupId             string                     `json:"groupId" xml:"groupId"`
	Fields              []ExportScrapeResultDetail `json:"fields" xml:"fields>field"`
	TimestampInitial    string                     `json:"timestampInitial" xml:"timestampInitial"`
	TimestampLastUpdate string                     `json:"timestampLastUpdate" xml:"timestampLastUpdate"`
	GroupVersionTag     string                     `json:"groupVersionTag" xml:"groupVersionTag"`
}

type ScrapeResultDetailTest struct {
	ID      string      `json:"id"`
	FieldID string      `json:"fieldId"`
	Value   interface{} `json:"value"`
	RawData string      `json:"rawData"`
}

type FieldChangeType string

const (
	ChangeFieldKey  FieldChangeType = "change_field_key"
	ChangeFieldType FieldChangeType = "change_field_type"
	ChangeFieldName FieldChangeType = "change_field_name"
	DeleteField     FieldChangeType = "delete_field"
	AddField        FieldChangeType = "add_field"
)

type FieldChange struct {
	FieldID                 string          `json:"fieldId"`
	FieldIsNewSinceLastSave bool            `json:"fieldIsNewSinceLastSave"`
	ChangeType              FieldChangeType `json:"type"`
}
