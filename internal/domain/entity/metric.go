package entity

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metric) Valid() map[string]string {
	problems := make(map[string]string)
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			problems["value"] = "should not be empty"
		}
	case "counter":
		if m.Delta == nil {
			problems["delta"] = "should not be empty"
		}
	default:
		problems["type"] = "should not be empty"
	}
	if m.ID == "" {
		problems["id"] = "should not be empty"
	}
	return problems
}

type GetMetricDTO struct {
	MType string
	ID    string
}

func (dto GetMetricDTO) Valid() map[string]string {
	problems := make(map[string]string)
	if dto.MType == "" {
		problems["type"] = "should not be empty"
	}
	if dto.ID == "" {
		problems["id"] = "should not be empty"
	}
	return problems
}

type MetricSlices struct {
	Gauge   []Metric
	Counter []Metric
}
