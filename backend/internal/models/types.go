package models

import "time"

// ScrapeGroup represents a group of scrapes with associated fields and endpoints
type ScrapeGroup struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Fields    []Field    `json:"fields"`
	Endpoints []Endpoint `json:"endpoints"`
}

// Field represents a single field with its type and name
type Field struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
	Type FieldType `json:"type"`
}

// FieldType defines the type of a field
type FieldType string

const (
	FieldTypeText  FieldType = "text"
	FieldTypeImage FieldType = "image"
	FieldTypeLink  FieldType = "link"
)

// Endpoint represents an endpoint with its configuration
type Endpoint struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	URL                  string           `json:"url"`
	PaginationConfig     PaginationConfig `json:"paginationConfig"`
	MainElementSelector  string           `json:"mainElementSelector"`
	DetailFieldSelectors []FieldSelector  `json:"detailFieldSelectors"`
	Interval             *int             `json:"interval,omitempty"`
	Active               *bool            `json:"active,omitempty"`
	LastScraped          *time.Time       `json:"lastScraped,omitempty"`
	Status               *ScrapeStatus    `json:"status,omitempty"`
}

// PaginationConfig represents pagination configuration for an endpoint
type PaginationConfig struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter,omitempty"`
	Start     int    `json:"start,omitempty"`
	End       int    `json:"end,omitempty"`
}

// ScrapeStatus defines the status of a scrape
type ScrapeStatus string

const (
	ScrapeStatusSuccess ScrapeStatus = "success"
	ScrapeStatusFailed  ScrapeStatus = "failed"
)

// FieldSelector represents a selector for a field
type FieldSelector struct {
	ID             string `json:"id"`
	FieldID        string `json:"fieldId"`
	Selector       string `json:"selector"`
	AttributeToGet string `json:"attributeToGet"`
}

// SearchConfig represents search configuration for an endpoint
type SearchConfig struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Param string `json:"param"`
	Value string `json:"value"`
}

type FieldSelectorsResponse struct {
	Field          string `json:"field"`
	Selector       string `json:"selector"`
	Regex          string `json:"regex"`
	AttributeToGet string `json:"attributeToGet"`
}

type FieldSelectorsRequest struct {
	URL                         string   `json:"url"`
	MainElementSelector         string   `json:"mainElementSelector"`
	FieldsToExtractSelectorsFor []string `json:"fieldsToExtractSelectorsFor"`
}
