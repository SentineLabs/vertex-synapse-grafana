package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	opts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("http client options: %w", err)
	}

	// Create HTTP client with custom RoundTripper to add API key header
	opts.Middlewares = []httpclient.Middleware{}

	cl, err := httpclient.NewProvider().New(opts)
	if err != nil {
		return nil, fmt.Errorf("httpclient new: %w", err)
	}

	// Parse configuration
	var config Config
	if err := json.Unmarshal(settings.JSONData, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Get API key from secure JSON data
	apiKey := ""
	if settings.DecryptedSecureJSONData != nil {
		if val, exists := settings.DecryptedSecureJSONData["apiKey"]; exists {
			apiKey = val
		}
	}

	return &Datasource{
		httpClient: &httpClientWrapper{
			client: cl,
			apiKey: apiKey,
		},
		settings: settings,
		config:   config,
	}, nil
}

// Config holds the datasource configuration
type Config struct {
	Version       string `json:"version"`
	Timeout       int    `json:"timeout"`
	TLSSkipVerify bool   `json:"tlsSkipVerify"`
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	settings   backend.DataSourceInstanceSettings
	httpClient *httpClientWrapper
	config     Config
	queryModel *QueryModel // Store current query model for parsing functions
}

// httpClientWrapper wraps the HTTP client to add the API key header
type httpClientWrapper struct {
	client *http.Client
	apiKey string
}

// Do executes the HTTP request with the API key header
func (c *httpClientWrapper) Do(req *http.Request) (*http.Response, error) {
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}
	return c.client.Do(req)
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Parse the query
	var qm QueryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		response.Error = fmt.Errorf("unmarshal query: %w", err)
		return response
	}

	if qm.StormQuery == "" {
		response.Error = fmt.Errorf("storm query is required")
		return response
	}

	// Add Grafana time range to opts
	qm = d.injectTimeRange(qm, query.TimeRange)

	// Execute Storm query
	var frames data.Frames
	if qm.UseCall {
		frames, err = d.queryStormCall(ctx, qm, query.RefID)
	} else {
		frames, err = d.queryStorm(ctx, qm, query.RefID)
	}

	if err != nil {
		response.Error = err
		return response
	}
	response.Frames = frames

	return response
}

// QueryModel represents the query structure
type QueryModel struct {
	StormQuery string                 `json:"stormQuery"`
	UseCall    bool                   `json:"useCall"`
	Opts       map[string]interface{} `json:"opts"`
}

func (d *Datasource) injectTimeRange(qm QueryModel, timeRange backend.TimeRange) QueryModel {
	// Initialize opts if nil
	if qm.Opts == nil {
		qm.Opts = make(map[string]interface{})
	}

	// Initialize vars if not present
	vars, ok := qm.Opts["vars"].(map[string]interface{})
	if !ok || vars == nil {
		vars = make(map[string]interface{})
	}

	// Primary Storm time range variables
	// Storm expects ISO format strings for absolute time: YYYY-MM-DDTHH:MM:SS.sssZ
	vars["timeFrom"] = timeRange.From.Format("2006-01-02T15:04:05.000Z")
	vars["timeTo"] = timeRange.To.Format("2006-01-02T15:04:05.000Z")

	// Storm tuple format for @= range queries like .created@=($timeRange)
	// This is the primary format for Storm time range filtering
	vars["timeRange"] = []interface{}{
		timeRange.From.Format("2006-01-02T15:04:05.000Z"),
		timeRange.To.Format("2006-01-02T15:04:05.000Z"),
	}

	// Alternative formats for flexibility
	// Date only format (YYYY-MM-DD)
	vars["dateFrom"] = timeRange.From.Format("2006-01-02")
	vars["dateTo"] = timeRange.To.Format("2006-01-02")

	// Unix timestamps (milliseconds) - for custom Storm functions
	vars["timeFromMs"] = timeRange.From.UnixMilli()
	vars["timeToMs"] = timeRange.To.UnixMilli()

	// Unix timestamps (seconds)
	vars["timeFromSec"] = timeRange.From.Unix()
	vars["timeToSec"] = timeRange.To.Unix()

	// Update opts with the vars
	qm.Opts["vars"] = vars

	return qm
}

// StormMessage represents a message from the Storm API
type StormMessage []interface{}

// StormNode represents a node structure from Storm
type StormNode struct {
	Form  string
	Value interface{}
	Iden  string
	Tags  map[string]interface{}
	Props map[string]interface{}
}

func (d *Datasource) queryStorm(ctx context.Context, qm QueryModel, refID string) (data.Frames, error) {
	// Build request URL for Storm query
	url := fmt.Sprintf("%s/api/v1/storm", d.settings.URL)

	// Create request body with query and opts
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": qm.StormQuery,
		"opts":  qm.Opts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storm query failed with status: %d", resp.StatusCode)
	}

	// Parse streaming response - collect all nodes first
	type NodeRecord struct {
		Form  string
		Value string
		Iden  string
		Tags  string
		Props map[string]interface{}
	}
	var nodes []NodeRecord
	allPropKeys := make(map[string]bool)

	decoder := json.NewDecoder(resp.Body)
	for {
		var msg StormMessage
		if err := decoder.Decode(&msg); err != nil {
			if err.Error() == "EOF" || err.Error() == "unexpected end of JSON input" {
				break
			}
			// Try to continue on partial errors
			log.DefaultLogger.Warn("Error decoding storm message", "error", err)
			continue
		}

		if len(msg) < 2 {
			continue
		}

		msgType, ok := msg[0].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "node":
			// Parse node structure: ["node", [[form, value], {props}]]
			if nodeData, ok := msg[1].([]interface{}); ok && len(nodeData) >= 2 {
				node := NodeRecord{
					Props: make(map[string]interface{}),
				}

				if nodeDef, ok := nodeData[0].([]interface{}); ok && len(nodeDef) >= 2 {
					if form, ok := nodeDef[0].(string); ok {
						node.Form = form
						node.Value = fmt.Sprintf("%v", nodeDef[1])
					}
				}

				if nodeProps, ok := nodeData[1].(map[string]interface{}); ok {
					if iden, ok := nodeProps["iden"].(string); ok {
						node.Iden = iden
					}

					// Extract tags
					if nodeTags, ok := nodeProps["tags"].(map[string]interface{}); ok {
						var tagList []string
						for tag := range nodeTags {
							tagList = append(tagList, tag)
						}
						node.Tags = strings.Join(tagList, ", ")
					}

					// Extract all properties
					if props, ok := nodeProps["props"].(map[string]interface{}); ok {
						for k, v := range props {
							node.Props[k] = v
							allPropKeys[k] = true
						}
					}

					// Also add reprs if present for better readability
					if reprs, ok := nodeProps["reprs"].(map[string]interface{}); ok {
						for k, v := range reprs {
							// Store repr values with _repr suffix
							reprKey := k + "_repr"
							node.Props[reprKey] = v
							allPropKeys[reprKey] = true
						}
					}
				}

				nodes = append(nodes, node)
			}
		case "err":
			// Handle error message
			if errData, ok := msg[1].([]interface{}); ok && len(errData) >= 2 {
				return nil, fmt.Errorf("storm error: %v", errData[1])
			}
		case "fini":
			// Query finished
			goto done
		}
	}
done:

	// Build data frame from collected nodes
	frame := data.NewFrame("storm")
	frame.RefID = refID

	if len(nodes) > 0 {
		// Create base columns
		forms := make([]string, len(nodes))
		values := make([]string, len(nodes))
		idens := make([]string, len(nodes))
		tags := make([]string, len(nodes))

		for i, node := range nodes {
			forms[i] = node.Form
			values[i] = node.Value
			idens[i] = node.Iden
			tags[i] = node.Tags
		}

		frame.Fields = append(frame.Fields,
			data.NewField("form", nil, forms),
			data.NewField("value", nil, values),
			data.NewField("iden", nil, idens),
			data.NewField("tags", nil, tags),
		)

		// Create sorted list of property keys
		propKeys := make([]string, 0, len(allPropKeys))
		for k := range allPropKeys {
			propKeys = append(propKeys, k)
		}
		sort.Strings(propKeys)

		// Add a column for each property
		for _, propKey := range propKeys {
			// Check if this is a time field - be more inclusive
			lowerKey := strings.ToLower(propKey)
			isTimeField := strings.Contains(lowerKey, "created") ||
				strings.Contains(lowerKey, "seen") ||
				strings.Contains(lowerKey, "time") ||
				strings.Contains(lowerKey, "modified") ||
				strings.Contains(lowerKey, "updated") ||
				strings.Contains(lowerKey, "accessed") ||
				strings.Contains(lowerKey, "published") ||
				strings.Contains(lowerKey, "date") ||
				strings.Contains(lowerKey, "timestamp")

			// Skip _repr fields for time columns since we're formatting them properly
			if strings.HasSuffix(propKey, "_repr") && isTimeField {
				continue
			}

			if isTimeField && !strings.HasSuffix(propKey, "_repr") {
				// Handle as time field
				timeValues := make([]*time.Time, len(nodes))
				for i, node := range nodes {
					if val, exists := node.Props[propKey]; exists {
						if timeVal := d.parseTimeValue(val); timeVal != nil {
							timeValues[i] = timeVal
						}
					}
				}
				frame.Fields = append(frame.Fields,
					data.NewField(propKey, nil, timeValues),
				)
			} else {
				// Handle as string field
				propValues := make([]string, len(nodes))
				for i, node := range nodes {
					if val, exists := node.Props[propKey]; exists {
						propValues[i] = fmt.Sprintf("%v", val)
					} else {
						propValues[i] = ""
					}
				}
				frame.Fields = append(frame.Fields,
					data.NewField(propKey, nil, propValues),
				)
			}
		}
	}

	return data.Frames{frame}, nil
}

// parseTimeValueFromString attempts to parse a string value as time
func (d *Datasource) parseTimeValueFromString(val string) *time.Time {
	if val == "" {
		return nil
	}

	// Try to parse as number first (epoch time)
	if numVal, err := strconv.ParseFloat(val, 64); err == nil {
		if numVal > 1e12 && numVal < 2e12 {
			// Likely milliseconds
			t := time.Unix(0, int64(numVal)*1e6)
			return &t
		} else if numVal > 1e9 && numVal < 2e9 {
			// Likely seconds
			t := time.Unix(int64(numVal), 0)
			return &t
		}
	}

	// Try ISO formats
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02T15:04:05.000Z", val); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", val); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
		return &t
	}

	return nil
}

// parseTimeValue attempts to parse a value as a time
func (d *Datasource) parseTimeValue(val interface{}) *time.Time {
	switch v := val.(type) {
	case float64:
		// Synapse times are in milliseconds since epoch
		// Values like 1757289600 are Synapse milliseconds (not Unix seconds)
		if v > 1e9 {
			// Treat as milliseconds
			t := time.Unix(0, int64(v)*1e6)
			return &t
		}
	case int64:
		if v > 1e9 {
			// Treat as milliseconds
			t := time.Unix(0, v*1e6)
			return &t
		}
	case string:
		// Try to parse ISO format
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return &t
		}
		// Try other common formats
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", v); err == nil {
			return &t
		}
	}
	return nil
}

// flattenObject flattens a nested object into dot-notation keys
func (d *Datasource) flattenObject(obj map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, val := range obj {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := val.(type) {
		case map[string]interface{}:
			// Recursively flatten nested objects
			flattened := d.flattenObject(v, newKey)
			for k, fv := range flattened {
				result[k] = fv
			}
		case []interface{}:
			// Arrays are serialized as JSON
			jsonBytes, _ := json.Marshal(v)
			result[newKey] = string(jsonBytes)
		case float64, int, int64, bool:
			// Preserve numeric and boolean types
			result[newKey] = val
		case nil:
			result[newKey] = nil
		default:
			result[newKey] = fmt.Sprintf("%v", val)
		}
	}

	return result
}

// valueToString converts a value to string, serializing nested structures as JSON
func (d *Datasource) valueToString(val interface{}) string {
	switch v := val.(type) {
	case map[string]interface{}, []interface{}:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Store query model for access in parsing functions
var queryModel *QueryModel

func (d *Datasource) queryStormCall(ctx context.Context, qm QueryModel, refID string) (data.Frames, error) {
	// Store for access in parsing functions
	d.queryModel = &qm
	// Build request URL for Storm call
	url := fmt.Sprintf("%s/api/v1/storm/call", d.settings.URL)

	// Create request body with query and opts
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": qm.StormQuery,
		"opts":  qm.Opts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storm call failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Extract the actual result from the response
	if status, ok := response["status"].(string); ok && status == "ok" {
		if result, exists := response["result"]; exists {
			return d.parseStormCallResult(result, refID)
		}
	}

	// If no result field or status not ok, return the whole response
	return d.parseStormCallResult(response, refID)
}

func (d *Datasource) parseStormCallResult(result interface{}, refID string) (data.Frames, error) {
	frame := data.NewFrame("storm_call")
	frame.RefID = refID

	switch v := result.(type) {
	case []interface{}:
		// Handle list of items
		if len(v) == 0 {
			// Empty list
			frame.Fields = append(frame.Fields,
				data.NewField("result", nil, []string{}),
			)
			return data.Frames{frame}, nil
		}

		// Check first item to determine list type
		firstItem := v[0]
		switch firstItem.(type) {
		case map[string]interface{}:
			// List of objects - create table with columns from object keys
			return d.parseObjectList(v, refID)
		case []interface{}:
			// Could be list of nodes in [[form, value], {props}] format
			if isNodeList(v) {
				return d.parseNodeList(v, refID)
			}
			// Otherwise treat as list of lists
			return d.parseListOfLists(v, refID)
		default:
			// List of primitives
			return d.parsePrimitiveList(v, refID)
		}

	case map[string]interface{}:
		// Single object - create key/value table with proper type detection
		keys := make([]string, 0, len(v))
		values := make([]interface{}, 0, len(v))
		
		// Sort keys for consistent ordering
		sortedKeys := make([]string, 0, len(v))
		for k := range v {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		
		for _, k := range sortedKeys {
			keys = append(keys, k)
			val := v[k]
			values = append(values, val)
		}
		
		// Detect the type of values
		valueType := d.detectFieldType(values)
		
		frame.Fields = append(frame.Fields,
			data.NewField("key", nil, keys),
		)
		
		// Add value field with appropriate type
		switch valueType {
		case "float":
			floatValues := make([]*float64, len(values))
			for i, val := range values {
				if val != nil {
					switch v := val.(type) {
					case float64:
						floatValues[i] = &v
					case int:
						f := float64(v)
						floatValues[i] = &f
					case int64:
						f := float64(v)
						floatValues[i] = &f
					case string:
						// Try to parse string as number
						if numVal, err := strconv.ParseFloat(v, 64); err == nil {
							floatValues[i] = &numVal
						}
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField("value", nil, floatValues),
			)
		case "int":
			intValues := make([]*int64, len(values))
			for i, val := range values {
				if val != nil {
					switch v := val.(type) {
					case float64:
						intVal := int64(v)
						intValues[i] = &intVal
					case int:
						intVal := int64(v)
						intValues[i] = &intVal
					case int64:
						intValues[i] = &v
					case string:
						// Try to parse string as number
						if numVal, err := strconv.ParseInt(v, 10, 64); err == nil {
							intValues[i] = &numVal
						}
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField("value", nil, intValues),
			)
		case "bool":
			boolValues := make([]*bool, len(values))
			for i, val := range values {
				if val != nil {
					if b, ok := val.(bool); ok {
						boolValues[i] = &b
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField("value", nil, boolValues),
			)
		default:
			// String field
			stringValues := make([]string, len(values))
			for i, val := range values {
				if val != nil {
					stringValues[i] = fmt.Sprintf("%v", val)
				} else {
					stringValues[i] = ""
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField("value", nil, stringValues),
			)
		}
		
		return data.Frames{frame}, nil

	default:
		// Primitive value or null
		frame.Fields = append(frame.Fields,
			data.NewField("result", nil, []string{fmt.Sprintf("%v", result)}),
		)
		return data.Frames{frame}, nil
	}
}

func isNodeList(items []interface{}) bool {
	if len(items) == 0 {
		return false
	}
	// Check if first item looks like a node: [[form, value], {props}]
	if nodeData, ok := items[0].([]interface{}); ok && len(nodeData) >= 2 {
		if nodeDef, ok := nodeData[0].([]interface{}); ok && len(nodeDef) >= 2 {
			if _, ok := nodeDef[0].(string); ok {
				return true
			}
		}
	}
	return false
}

func (d *Datasource) parseNodeList(items []interface{}, refID string) (data.Frames, error) {
	frame := data.NewFrame("storm_call")
	frame.RefID = refID

	var forms []string
	var values []string
	var idens []string
	var tags []string

	for _, item := range items {
		if nodeData, ok := item.([]interface{}); ok && len(nodeData) >= 2 {
			if nodeDef, ok := nodeData[0].([]interface{}); ok && len(nodeDef) >= 2 {
				if form, ok := nodeDef[0].(string); ok {
					forms = append(forms, form)
					values = append(values, fmt.Sprintf("%v", nodeDef[1]))
				}
			}
			if nodeProps, ok := nodeData[1].(map[string]interface{}); ok {
				if iden, ok := nodeProps["iden"].(string); ok {
					idens = append(idens, iden)
				} else {
					idens = append(idens, "")
				}
				// Extract tags if present
				if nodeTags, ok := nodeProps["tags"].(map[string]interface{}); ok {
					var tagList []string
					for tag := range nodeTags {
						tagList = append(tagList, tag)
					}
					tags = append(tags, strings.Join(tagList, ", "))
				} else {
					tags = append(tags, "")
				}
			}
		}
	}

	if len(forms) > 0 {
		frame.Fields = append(frame.Fields,
			data.NewField("form", nil, forms),
			data.NewField("value", nil, values),
			data.NewField("iden", nil, idens),
			data.NewField("tags", nil, tags),
		)
	}

	return data.Frames{frame}, nil
}

func (d *Datasource) parseObjectList(items []interface{}, refID string) (data.Frames, error) {
	frame := data.NewFrame("storm_call")
	frame.RefID = refID

	if len(items) == 0 {
		return data.Frames{frame}, nil
	}

	// Check if we should flatten nested objects
	shouldFlatten := false
	if d.queryModel != nil && d.queryModel.Opts != nil {
		if flatten, ok := d.queryModel.Opts["flatten"].(bool); ok {
			shouldFlatten = flatten
		}
	}

	// Get all unique keys from all objects
	keySet := make(map[string]bool)
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok {
			if shouldFlatten {
				// Collect flattened keys
				flattened := d.flattenObject(obj, "")
				for k := range flattened {
					keySet[k] = true
				}
			} else {
				for k := range obj {
					keySet[k] = true
				}
			}
		}
	}

	// Create sorted list of keys
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create fields for each key - use interface{} to preserve types
	fields := make(map[string][]interface{})
	for _, key := range keys {
		fields[key] = []interface{}{}
	}

	// Populate field values
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok {
			if shouldFlatten {
				// Flatten the object preserving types
				flattened := d.flattenObject(obj, "")
				for _, key := range keys {
					if val, exists := flattened[key]; exists {
						fields[key] = append(fields[key], val)
					} else {
						fields[key] = append(fields[key], nil)
					}
				}
			} else {
				// Preserve types in non-flattened mode too
				for _, key := range keys {
					if val, exists := obj[key]; exists {
						// Only convert nested structures to JSON, preserve primitive types
						switch v := val.(type) {
						case map[string]interface{}, []interface{}:
							jsonStr := d.valueToString(v)
							fields[key] = append(fields[key], jsonStr)
						default:
							fields[key] = append(fields[key], val)
						}
					} else {
						fields[key] = append(fields[key], nil)
					}
				}
			}
		}
	}

	// Add fields to frame with type detection
	for _, key := range keys {
		// Determine field type from values
		fieldType := d.detectFieldType(fields[key])

		// Check if this is a time field
		lowerKey := strings.ToLower(key)
		isTimeField := strings.Contains(lowerKey, "created") ||
			strings.Contains(lowerKey, "seen") ||
			strings.Contains(lowerKey, "time") ||
			strings.Contains(lowerKey, "modified") ||
			strings.Contains(lowerKey, "updated") ||
			strings.Contains(lowerKey, "accessed") ||
			strings.Contains(lowerKey, "published") ||
			strings.Contains(lowerKey, "date") ||
			strings.Contains(lowerKey, "timestamp")

		if isTimeField && (fieldType == "float" || fieldType == "int") {
			// Try to parse numeric values as timestamps
			timeValues := make([]*time.Time, len(fields[key]))
			hasTimeValues := false
			for i, val := range fields[key] {
				if timeVal := d.parseTimeValue(val); timeVal != nil {
					timeValues[i] = timeVal
					hasTimeValues = true
				}
			}
			if hasTimeValues {
				frame.Fields = append(frame.Fields,
					data.NewField(key, nil, timeValues),
				)
				continue
			}
		}

		// Add field based on detected type
		switch fieldType {
		case "float":
			floatValues := make([]*float64, len(fields[key]))
			for i, val := range fields[key] {
				if val != nil {
					switch v := val.(type) {
					case float64:
						floatValues[i] = &v
					case int:
						f := float64(v)
						floatValues[i] = &f
					case int64:
						f := float64(v)
						floatValues[i] = &f
					case string:
						// Try to parse string as number
						if numVal, err := strconv.ParseFloat(v, 64); err == nil {
							floatValues[i] = &numVal
						}
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField(key, nil, floatValues),
			)
		case "int":
			intValues := make([]*int64, len(fields[key]))
			for i, val := range fields[key] {
				if val != nil {
					switch v := val.(type) {
					case float64:
						intVal := int64(v)
						intValues[i] = &intVal
					case int:
						intVal := int64(v)
						intValues[i] = &intVal
					case int64:
						intValues[i] = &v
					case string:
						// Try to parse string as number
						if numVal, err := strconv.ParseInt(v, 10, 64); err == nil {
							intValues[i] = &numVal
						}
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField(key, nil, intValues),
			)
		case "bool":
			boolValues := make([]*bool, len(fields[key]))
			for i, val := range fields[key] {
				if val != nil {
					if b, ok := val.(bool); ok {
						boolValues[i] = &b
					}
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField(key, nil, boolValues),
			)
		default:
			// String field
			stringValues := make([]string, len(fields[key]))
			for i, val := range fields[key] {
				if val != nil {
					stringValues[i] = fmt.Sprintf("%v", val)
				} else {
					stringValues[i] = ""
				}
			}
			frame.Fields = append(frame.Fields,
				data.NewField(key, nil, stringValues),
			)
		}
	}

	return data.Frames{frame}, nil
}

// detectFieldType detects the common type of values in a slice
func (d *Datasource) detectFieldType(values []interface{}) string {
	if len(values) == 0 {
		return "string"
	}

	hasFloat := false
	hasInt := false
	hasBool := false
	hasOther := false

	// Scan all non-nil values to determine the best type
	for _, val := range values {
		if val == nil {
			continue
		}

		switch v := val.(type) {
		case float64:
			// Check if it's a whole number (integer)
			if v == float64(int64(v)) {
				hasInt = true
			} else {
				hasFloat = true
			}
		case int, int64:
			hasInt = true
		case bool:
			hasBool = true
		case string:
			// Try to parse as number
			if numVal, err := strconv.ParseFloat(v, 64); err == nil {
				if numVal == float64(int64(numVal)) {
					hasInt = true
				} else {
					hasFloat = true
				}
			} else {
				hasOther = true
			}
		default:
			hasOther = true
		}
	}

	// Determine the best type based on what we found
	if hasOther {
		return "string"
	}
	if hasBool && !hasInt && !hasFloat {
		return "bool"
	}
	if hasFloat {
		return "float"
	}
	if hasInt {
		return "int"
	}

	return "string"
}

func (d *Datasource) parsePrimitiveList(items []interface{}, refID string) (data.Frames, error) {
	frame := data.NewFrame("storm_call")
	frame.RefID = refID

	// Check if all items are potential timestamps
	allTimestamps := true
	timeValues := make([]*time.Time, len(items))
	for i, item := range items {
		// Try to parse as timestamp
		if timeVal := d.parseTimeValue(item); timeVal != nil {
			timeValues[i] = timeVal
		} else {
			allTimestamps = false
			break
		}
	}

	if allTimestamps && len(items) > 0 {
		// All values are timestamps - return as time field
		frame.Fields = append(frame.Fields,
			data.NewField("value", nil, timeValues),
		)
	} else {
		// Fall back to string values
		values := make([]string, len(items))
		for i, item := range items {
			values[i] = fmt.Sprintf("%v", item)
		}
		frame.Fields = append(frame.Fields,
			data.NewField("value", nil, values),
		)
	}

	return data.Frames{frame}, nil
}

func (d *Datasource) parseListOfLists(items []interface{}, refID string) (data.Frames, error) {
	frame := data.NewFrame("storm_call")
	frame.RefID = refID

	// Convert to JSON string representation for complex nested structures
	values := make([]string, len(items))
	for i, item := range items {
		jsonBytes, _ := json.Marshal(item)
		values[i] = string(jsonBytes)
	}

	frame.Fields = append(frame.Fields,
		data.NewField("value", nil, values),
	)

	return data.Frames{frame}, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called")

	status := backend.HealthStatusOk
	message := "Data source is working"

	// Test connection to Cortex API using Storm endpoint
	url := fmt.Sprintf("%s/api/v1/storm", d.settings.URL)
	reqBody := []byte(`{"query": ""}`)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		status = backend.HealthStatusError
		message = fmt.Sprintf("Failed to create request: %v", err)
		return &backend.CheckHealthResult{
			Status:  status,
			Message: message,
		}, nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		status = backend.HealthStatusError
		message = fmt.Sprintf("Failed to connect to Cortex: %v", err)
		return &backend.CheckHealthResult{
			Status:  status,
			Message: message,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status = backend.HealthStatusError
		message = fmt.Sprintf("Cortex returned status: %d", resp.StatusCode)
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
