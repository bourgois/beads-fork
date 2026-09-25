package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An unset must not also rewrite the end of the file. commentOutYamlKey used to
// drop the document's trailing newline, so `bd config unset` produced a
// no-newline-at-end-of-file change on top of the line it meant to comment out —
// and did so even for a key the document does not contain, i.e. when nothing
// was edited at all. A config.yaml is git-tracked, so that is a spurious line
// in someone's review.
func TestCommentOutYamlKeyPreservesTheTrailingNewline(t *testing.T) {
	for _, tc := range []struct {
		name, content, key string
	}{
		{"nested key present", "# header\nissue_prefix: vp\ndolt:\n  mode: server\n", "dolt.mode"},
		{"flat dotted key present", "issue_prefix: vp\ndolt.mode: server\n", "dolt.mode"},
		{"single-segment key present", "issue_prefix: vp\nexport.auto: true\n", "issue_prefix"},
		{"key absent — nothing is edited", "issue_prefix: vp\n", "dolt.mode"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := commentOutYamlKey(tc.content, tc.key)
			if err != nil {
				t.Fatalf("commentOutYamlKey: %v", err)
			}
			if !strings.HasSuffix(out, "\n") {
				t.Errorf("trailing newline dropped:\nin:  %q\nout: %q", tc.content, out)
			}
		})
	}

	// A document that genuinely has no trailing newline keeps not having one:
	// the rule is "preserve", not "always append".
	t.Run("absent trailing newline is not invented", func(t *testing.T) {
		out, err := commentOutYamlKey("issue_prefix: vp\ndolt.mode: server", "dolt.mode")
		if err != nil {
			t.Fatalf("commentOutYamlKey: %v", err)
		}
		if strings.HasSuffix(out, "\n") {
			t.Errorf("a trailing newline was invented for a document that had none: %q", out)
		}
	})
}

// The same property through the public writer, which is what `bd config unset`
// actually calls.
func TestUnsetThroughTheFileKeepsTheTrailingNewline(t *testing.T) {
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(beadsDir, "config.yaml")
	const body = "issue_prefix: vp\nexport.auto: true\ndolt:\n  mode: server\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_DIR", beadsDir)
	if err := UnsetYamlConfig("dolt.mode"); err != nil {
		t.Fatalf("UnsetYamlConfig: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(after), "\n") {
		t.Errorf("unset left the file without its trailing newline:\n%q", string(after))
	}
}
