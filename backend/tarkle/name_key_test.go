package tarkle

import (
	"context"
	"testing"

	"github.com/rclone/rclone/backend/tarkle/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tarkle compares names by NFC, one space for each run of white space, trim
// and lower case (normalizeName in lib/file-version-core.mjs). The backend
// has to find "the file that blocks this name" the same way, or a replace
// leaves the old file in place and the rename fails with name_taken.
func TestNameKeyIsTheKeyTarkleUses(t *testing.T) {
	for _, pair := range [][2]string{
		{"MEDIUM.bin", "medium.bin"},
		{"Café.txt", "Café.txt"}, // precomposed é and e + combining accent (macOS writes the second)
		{"a  b.txt", "a b.txt"},
		{"a\tb.txt", "a b.txt"},
		{" a.txt ", "a.txt"},
		{"a b.txt", "a b.txt"}, // no-break space
	} {
		assert.Equal(t, nameKey(pair[0]), nameKey(pair[1]), "%q and %q are the same name to Tarkle", pair[0], pair[1])
	}
	for _, pair := range [][2]string{
		{"a.txt", "b.txt"},
		{"a b.txt", "ab.txt"},
		{"a.txt", "a.txt.bak"},
	} {
		assert.NotEqual(t, nameKey(pair[0]), nameKey(pair[1]), "%q and %q are different names", pair[0], pair[1])
	}
}

func TestReplaceExistingFindsTheSameNameInAnotherForm(t *testing.T) {
	// the old file is stored decomposed and with a doubled space; the new one is asked to take the composed, single-space name
	f, s := newScriptServer(t, nil, replaceAnswers(fileJSON("old", "Café  menu.txt")+","+fileJSON("new", "Café menu (2).txt"), nil))
	item := &api.Item{ID: "new", Type: api.ItemTypeFile, Name: "Café menu (2).txt"}
	require.NoError(t, f.replaceExisting(context.Background(), "", "Café menu.txt", item))
	assert.Equal(t, []string{"PATCH /files/old", "PATCH /files/new", "DELETE /files/old"}, s.writes(), "the old file is found, set aside and binned after the swap")
}
