package models

import (
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
}

func (sg ScrapeGroup) GetEndpointById(id string) *Endpoint {
	for _, endpoint := range sg.Endpoints {
		if endpoint.ID == id {
			return &endpoint
		}
	}
	return nil
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
	ID                   string           `json:"id" bson:"id"`
	Name                 string           `json:"name" bson:"name"`
	URL                  string           `json:"url" bson:"url"`
	PaginationConfig     PaginationConfig `json:"paginationConfig" bson:"paginationConfig"`
	MainElementSelector  string           `json:"mainElementSelector" bson:"mainElementSelector"`
	DetailFieldSelectors []FieldSelector  `json:"detailFieldSelectors" bson:"detailFieldSelectors"`
	Interval             *int             `json:"interval,omitempty" bson:"interval,omitempty"`
	Active               *bool            `json:"active,omitempty" bson:"active,omitempty"`
	LastScraped          *time.Time       `json:"lastScraped,omitempty" bson:"lastScraped,omitempty"`
	Status               *ScrapeStatus    `json:"status,omitempty" bson:"status,omitempty"`
}

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
	ScrapeStatusSuccess ScrapeStatus = "success"
	ScrapeStatusFailed  ScrapeStatus = "failed"
)

// FieldSelector represents a selector for a field
type FieldSelector struct {
	ID             string `json:"id" bson:"id"`
	FieldID        string `json:"fieldId" bson:"fieldId"`
	Selector       string `json:"selector" bson:"selector"`
	Regex          string `json:"regex" bson:"regex"`
	AttributeToGet string `json:"attributeToGet" bson:"attributeToGet"`
}

// SearchConfig represents search configuration for an endpoint
type SearchConfig struct {
	ID    string `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Param string `json:"param" bson:"param"`
	Value string `json:"value" bson:"value"`
}

type FieldSelectorsResponse struct {
	Field          string `json:"field" bson:"field"`
	Selector       string `json:"selector" bson:"selector"`
	Regex          string `json:"regex" bson:"regex"`
	AttributeToGet string `json:"attributeToGet" bson:"attributeToGet"`
}

type FieldSelectorsRequest struct {
	URL                         string   `json:"url" bson:"url"`
	MainElementSelector         string   `json:"mainElementSelector" bson:"mainElementSelector"`
	FieldsToExtractSelectorsFor []string `json:"fieldsToExtractSelectorsFor" bson:"fieldsToExtractSelectorsFor"`
}

type ScrapeResult struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UniqueHash string               `json:"uniqueHash" bson:"uniqueHash"`
	EndpointID string               `json:"endpointId" bson:"endpointId"`
	GroupId    primitive.ObjectID   `json:"groupId" bson:"groupId"`
	Fields     []ScrapeResultDetail `json:"fields" bson:"fields"`
	Timestamp  string               `json:"timestamp" bson:"timestamp"`
}

type ScrapeResultDetail struct {
	ID      string `json:"id" bson:"id"`
	FieldID string `json:"fieldId" bson:"fieldId"`
	Value   string `json:"value" bson:"value"`
}
