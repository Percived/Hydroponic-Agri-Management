package metric

// MetricDefinitionResponse ...
type MetricDefinitionResponse struct {
	ID              uint64   `json:"id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Unit            string   `json:"unit"`
	PrecisionDigits uint8    `json:"precision_digits"`
	NormalRangeMin  *float64 `json:"normal_range_min"`
	NormalRangeMax  *float64 `json:"normal_range_max"`
	IsCore          uint8    `json:"is_core"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// MetricListResponse ...
type MetricListResponse struct {
	Items    []MetricDefinitionResponse `json:"items"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}
