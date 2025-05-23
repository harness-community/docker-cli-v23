package cli

import (
	"testing"

	pluginmanager "github.com/harness-community/docker-cli-v23/cli-plugins/manager"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestVisitAll(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub1 := &cobra.Command{Use: "sub1"}
	sub1sub1 := &cobra.Command{Use: "sub1sub1"}
	sub1sub2 := &cobra.Command{Use: "sub1sub2"}
	sub2 := &cobra.Command{Use: "sub2"}

	root.AddCommand(sub1, sub2)
	sub1.AddCommand(sub1sub1, sub1sub2)

	// Take the opportunity to test DisableFlagsInUseLine too
	DisableFlagsInUseLine(root)

	var visited []string
	VisitAll(root, func(ccmd *cobra.Command) {
		visited = append(visited, ccmd.Name())
		assert.Assert(t, ccmd.DisableFlagsInUseLine, "DisableFlagsInUseLine not set on %q", ccmd.Name())
	})
	expected := []string{"sub1sub1", "sub1sub2", "sub1", "sub2", "root"}
	assert.DeepEqual(t, expected, visited)
}

func TestVendorAndVersion(t *testing.T) {
	// Non plugin.
	assert.Equal(t, vendorAndVersion(&cobra.Command{Use: "test"}), "")

	// Plugins with various lengths of vendor.
	for _, tc := range []struct {
		vendor   string
		version  string
		expected string
	}{
		{vendor: "vendor", expected: "(vendor)"},
		{vendor: "vendor", version: "testing", expected: "(vendor, testing)"},
	} {
		t.Run(tc.vendor, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
				Annotations: map[string]string{
					pluginmanager.CommandAnnotationPlugin:        "true",
					pluginmanager.CommandAnnotationPluginVendor:  tc.vendor,
					pluginmanager.CommandAnnotationPluginVersion: tc.version,
				},
			}
			assert.Equal(t, vendorAndVersion(cmd), tc.expected)
		})
	}
}

func TestInvalidPlugin(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub1 := &cobra.Command{Use: "sub1"}
	sub1sub1 := &cobra.Command{Use: "sub1sub1"}
	sub1sub2 := &cobra.Command{Use: "sub1sub2"}
	sub2 := &cobra.Command{Use: "sub2"}

	assert.Assert(t, is.Len(invalidPlugins(root), 0))

	sub1.Annotations = map[string]string{
		pluginmanager.CommandAnnotationPlugin:        "true",
		pluginmanager.CommandAnnotationPluginInvalid: "foo",
	}
	root.AddCommand(sub1, sub2)
	sub1.AddCommand(sub1sub1, sub1sub2)

	assert.DeepEqual(t, invalidPlugins(root), []*cobra.Command{sub1}, cmpopts.IgnoreUnexported(cobra.Command{}))
}

func TestCommandAliases(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{Use: "subcommand", Aliases: []string{"alias1", "alias2"}}
	sub2 := &cobra.Command{Use: "subcommand2", Annotations: map[string]string{"aliases": "root foo, root bar"}}
	root.AddCommand(sub)
	root.AddCommand(sub2)

	assert.Equal(t, hasAliases(sub), true)
	assert.Equal(t, commandAliases(sub), "root subcommand, root alias1, root alias2")
	assert.Equal(t, hasAliases(sub2), true)
	assert.Equal(t, commandAliases(sub2), "root foo, root bar")

	sub.Annotations = map[string]string{"aliases": "custom alias, custom alias 2"}
	assert.Equal(t, hasAliases(sub), true)
	assert.Equal(t, commandAliases(sub), "custom alias, custom alias 2")
}

func TestDecoratedName(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	topLevelCommand := &cobra.Command{Use: "pluginTopLevelCommand"}
	root.AddCommand(topLevelCommand)
	assert.Equal(t, decoratedName(topLevelCommand), "pluginTopLevelCommand ")
	topLevelCommand.Annotations = map[string]string{pluginmanager.CommandAnnotationPlugin: "true"}
	assert.Equal(t, decoratedName(topLevelCommand), "pluginTopLevelCommand*")
}
