// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package status

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestKeyMsgForCompat(t *testing.T) {
	t.Parallel()

	msg, ok := KeyMsgFor([]string{"", "ctrl+s"})
	if !ok {
		t.Fatal("expected KeyMsgFor to parse ctrl+s")
	}
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		t.Fatalf("unexpected msg type %T", msg)
	}
	if press.Code != 's' || press.Mod != tea.ModCtrl {
		t.Fatalf("unexpected key press: %+v", press)
	}

	if _, ok := KeyMsgFor([]string{"not-a-key"}); ok {
		t.Fatal("expected unknown key spec to fail")
	}
}
