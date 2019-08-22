package atlas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLabel(t *testing.T) {
	labelled := Labelled{
		[]Label{
			Label{"key", "value"},
		},
	}

	assert.Equal(t, "value", labelled.GetLabel("key"))
}

func TestGetNonexistentLabel(t *testing.T) {
	labelled := Labelled{
		[]Label{},
	}

	assert.Equal(t, "", labelled.GetLabel("key"))
}

func TestSetLabel(t *testing.T) {
	labelled := Labelled{
		[]Label{},
	}

	labelled.SetLabel("key", "value")

	assert.Equal(t, []Label{Label{"key", "value"}}, labelled.Labels)
}

func TestSetExistingLabel(t *testing.T) {
	labelled := Labelled{
		[]Label{
			Label{"key", "value"},
		},
	}

	labelled.SetLabel("key", "value")

	assert.Equal(t, []Label{Label{"key", "value"}}, labelled.Labels)
}
