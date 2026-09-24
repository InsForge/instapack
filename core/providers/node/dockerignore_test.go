package node

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestInstallDockerignoreContext(t *testing.T) {
	for _, tt := range []struct {
		name   string
		ignore string
		want   []string
	}{
		{"symlinked_manifest", "/prisma\n", []string{"package.json", "tools/demo/package.json", "tools/keep/package.json", "linked/package.json", "config/private.txt"}},
		{"wildcard_reincluded_child", "/prisma\n!**/schema.prisma\n", []string{"package.json", "tools/demo/package.json", "tools/keep/package.json", "prisma", "config/private.txt"}},
		{"reincluded_child", "/prisma\n!/prisma/schema.prisma\n", []string{"package.json", "tools/demo/package.json", "tools/keep/package.json", "prisma", "config/private.txt"}},
		{"unfiltered", "", []string{"package.json", "tools/demo/package.json", "tools/keep/package.json", "prisma", "config/private.txt"}},
		{"excluded", "/tools/demo\n/prisma\n/config/private.txt\n", []string{"package.json", "tools/keep/package.json"}},
		{"reincluded", "/tools/*\n!/tools/keep\n/prisma\n/config/private.txt\n", []string{"package.json", "tools/keep/package.json"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, name := range []string{"package.json", "tools/demo/package.json", "tools/keep/package.json", "prisma/schema.prisma", "config/private.txt"} {
				path := filepath.Join(directory, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))
			}
			if tt.ignore != "" {
				require.NoError(t, os.WriteFile(filepath.Join(directory, ".dockerignore"), []byte(tt.ignore), 0o644))
			}
			if tt.name == "symlinked_manifest" {
				require.NoError(t, os.Symlink("tools/keep", filepath.Join(directory, "linked")))
			}
			userApp, err := app.NewApp(directory)
			require.NoError(t, err)
			ctx := &generate.GenerateContext{
				App:    userApp,
				Env:    app.NewEnvironment(nil),
				Logger: logger.NewLogger(),
				Caches: generate.NewCacheContext(),
			}
			ctx.Env.Variables["RAILPACK_NODE_INSTALL_PATTERNS"] = "config/*.txt"
			workspace, err := NewWorkspace(ctx.App)
			require.NoError(t, err)
			install := ctx.NewCommandStep("install")
			PackageManagerNpm.installDependencies(ctx, workspace, install, false)
			var copied []string
			for _, command := range install.Commands {
				if copy, ok := command.(plan.CopyCommand); ok {
					copied = append(copied, copy.Src)
				}
			}
			require.ElementsMatch(t, tt.want, copied)
		})
	}
}
