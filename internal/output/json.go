package output

import (
	"encoding/json"
	"io"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

// JSONFormatter outputs data as JSON
type JSONFormatter struct{}

func (f *JSONFormatter) FormatItems(w io.Writer, items []api.Item) error {
	return json.NewEncoder(w).Encode(items)
}

func (f *JSONFormatter) FormatItem(w io.Writer, item *api.Item) error {
	return json.NewEncoder(w).Encode(item)
}

func (f *JSONFormatter) FormatInstances(w io.Writer, instances []api.Instance) error {
	return json.NewEncoder(w).Encode(instances)
}

func (f *JSONFormatter) FormatInstance(w io.Writer, instance *api.Instance) error {
	return json.NewEncoder(w).Encode(instance)
}

func (f *JSONFormatter) FormatContext(w io.Writer, item *api.Item, instances []api.Instance) error {
	data := struct {
		Item      *api.Item      `json:"item"`
		Instances []api.Instance `json:"instances"`
	}{
		Item:      item,
		Instances: instances,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *JSONFormatter) FormatProjectInfo(w io.Writer, info *api.ProjectInfo) error {
	return json.NewEncoder(w).Encode(info)
}
