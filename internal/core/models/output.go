package models

import (
	"time"
)

// OutputData represents the final JSON output structure
type OutputData struct {
	Metadata OutputMetadata         `json:"metadata"`
	Schema   map[string][]FieldInfo `json:"schema"`
	Data     map[string][]interface{} `json:"data"`
}

// OutputMetadata contains metadata about the generated output
type OutputMetadata struct {
	GeneratedAt   time.Time `json:"generated_at"`
	SchemaVersion string    `json:"schema_version"`
	Generator     string    `json:"generator"`
	FileCount     int       `json:"file_count"`
	RecordCount   int       `json:"record_count"`
}

// FieldInfo represents schema information for a field
type FieldInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

// DataRecord represents a single data record
type DataRecord map[string]interface{}

// NewOutputData creates a new OutputData instance
func NewOutputData() *OutputData {
	return &OutputData{
		Metadata: OutputMetadata{
			GeneratedAt: time.Now(),
			Generator:   "Excel Schema Generator v2.0",
		},
		Schema: make(map[string][]FieldInfo),
		Data:   make(map[string][]interface{}),
	}
}

// SetMetadata sets the output metadata
func (o *OutputData) SetMetadata(fileCount, recordCount int, schemaVersion string) {
	o.Metadata.FileCount = fileCount
	o.Metadata.RecordCount = recordCount
	o.Metadata.SchemaVersion = schemaVersion
	o.Metadata.GeneratedAt = time.Now()
}

// AddSchema adds schema information for a class
func (o *OutputData) AddSchema(className string, fields []FieldInfo) {
	o.Schema[className] = fields
}

// AddData adds data records for a class
func (o *OutputData) AddData(className string, records []interface{}) {
	o.Data[className] = records
}

// GetClassCount returns the number of classes in the output
func (o *OutputData) GetClassCount() int {
	return len(o.Schema)
}

// GetTotalRecordCount returns the total number of records across all classes
func (o *OutputData) GetTotalRecordCount() int {
	total := 0
	for _, records := range o.Data {
		total += len(records)
	}
	return total
}

// HasClass checks if a class exists in the output
func (o *OutputData) HasClass(className string) bool {
	_, exists := o.Schema[className]
	return exists
}

// GetSchema returns schema information for a class
func (o *OutputData) GetSchema(className string) ([]FieldInfo, bool) {
	schema, exists := o.Schema[className]
	return schema, exists
}

// GetData returns data records for a class
func (o *OutputData) GetData(className string) ([]interface{}, bool) {
	data, exists := o.Data[className]
	return data, exists
}

// GetClassNames returns all class names
func (o *OutputData) GetClassNames() []string {
	names := make([]string, 0, len(o.Schema))
	for name := range o.Schema {
		names = append(names, name)
	}
	return names
}

// NewFieldInfo creates a new FieldInfo instance
func NewFieldInfo(name, dataType string) FieldInfo {
	return FieldInfo{
		Name:     name,
		DataType: dataType,
	}
}

// NewDataRecord creates a new DataRecord instance
func NewDataRecord() DataRecord {
	return make(DataRecord)
}

// Set sets a field value in the data record
func (r DataRecord) Set(key string, value interface{}) {
	r[key] = value
}

// Get retrieves a field value from the data record
func (r DataRecord) Get(key string) (interface{}, bool) {
	value, exists := r[key]
	return value, exists
}

// Has checks if a field exists in the data record
func (r DataRecord) Has(key string) bool {
	_, exists := r[key]
	return exists
}

// Keys returns all keys in the data record
func (r DataRecord) Keys() []string {
	keys := make([]string, 0, len(r))
	for key := range r {
		keys = append(keys, key)
	}
	return keys
}