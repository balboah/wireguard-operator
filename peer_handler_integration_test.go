// +build integration

package operator

import (
	"testing"
)

func TestPeerHandlerIntegration(t *testing.T) {
	wg0, err := NewWgLink("wgo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { wg0.Close() }()
	c, err := NewWgClient(wg0, 1234, "")
	if err != nil {
		t.Fatal(err)
	}

	testPeerHandler(t, c)
}
