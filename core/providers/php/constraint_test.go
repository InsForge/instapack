package php

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestComposerPhpConstraint(t *testing.T) {
	for _, tc := range []struct{ constraint, requested string }{
		{"^8.2.0", "8.2"},
		{"^8.3.0", "8.3"},
		{"^8.2.7", "8.2.7"},
		{"8.2.0", "8.2.0"},
		{"^8.2", "8.2"},
		{"^8.2.0-beta.1", "8.2.0-beta.1"},
		{"^8.2.0+build", "8.2.0+build"},
		{"^0.0.0", "0.0.0"},
		{"^8.2.0|^8.3", "8.2.0|^8.3"},
		{"^8.2.0@dev", "8.2.0@dev"},
		{">=8.2 <8.4", ">=8.2 <8.4"},
	} {
		t.Run(tc.constraint, func(t *testing.T) {
			directory := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(directory, "composer.json"), []byte(`{"require":{"php":"`+tc.constraint+`"}}`), 0600))
			ctx := testingUtils.CreateGenerateContext(t, directory)
			provider := PhpProvider{}
			_, err := provider.phpImagePackage(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.requested, ctx.Resolver.Get("php").Version)
		})
	}
}
