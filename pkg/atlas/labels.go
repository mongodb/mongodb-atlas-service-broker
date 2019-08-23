package atlas

// Labelled is meant to be embedded in Atlas models which have labels. It adds
// helper function to get and set labels.
type Labelled struct {
	Labels []Label `json:"labels,omitempty"`
}

// Label represents a single label on an Atlas object consisting of a key-value
// pair.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetLabel will find a label by key and return its value.
func (l Labelled) GetLabel(key string) string {
	for _, label := range l.Labels {
		if label.Key == key {
			return label.Value
		}
	}

	return ""
}

// SetLabel will update or add a new label with the specified key and value.
func (l *Labelled) SetLabel(key, value string) {
	// Check if label already exists, if so update it.
	for _, label := range l.Labels {
		if label.Key == key {
			label.Value = value
			return
		}
	}

	l.Labels = append(l.Labels, Label{Key: key, Value: value})
}
