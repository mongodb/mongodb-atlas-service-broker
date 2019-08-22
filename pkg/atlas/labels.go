package atlas

type Labelled struct {
	Labels []Label `json:"labels,omitempty"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (l Labelled) GetLabel(key string) string {
	for _, label := range l.Labels {
		if label.Key == key {
			return label.Value
		}
	}

	return ""
}

func (l *Labelled) SetLabel(key, value string) {
	for _, label := range l.Labels {
		if label.Key == key {
			label.Value = value
			return
		}
	}

	l.Labels = append(l.Labels, Label{Key: key, Value: value})
}
