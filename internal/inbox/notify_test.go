package inbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendNotification_DoesNotPanic(t *testing.T) {
	// SendNotification should not panic regardless of platform or available tools.
	// On CI or systems without notification tools, it should gracefully handle
	// the absence and either succeed silently or return an error.
	assert.NotPanics(t, func() {
		_ = SendNotification("Guardian", "You have pending proposals")
	})
}

func TestSendNotification_EmptyArgs(t *testing.T) {
	// Should not panic with empty arguments
	assert.NotPanics(t, func() {
		_ = SendNotification("", "")
	})
}

func TestSendNotification_SpecialCharacters(t *testing.T) {
	// Should not panic with special characters in title/message
	assert.NotPanics(t, func() {
		_ = SendNotification("Guardian's Test", `He said "hello" & goodbye`)
	})
}
