package install

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFileMarkers(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)
	ioutil.WriteFile(path.Join(dir, "file-marker.nix"), []byte(`
  # File-marker: Trim on install
  { self, super }: {}`), 0644)

	t.Run("remove", func(t *testing.T) {
		if err := (&Applyer{}).applyFile(path.Join(dir, "file-marker.nix"), 0); err != nil {
			t.Error(err)
		}
		if _, err := os.Stat(path.Join(dir, "file-marker.nix")); !os.IsNotExist(err) {
			t.Errorf("wanted not found, got %v", err)
		}
	})

}

func TestInlineMarkers(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)
	ioutil.WriteFile(path.Join(dir, "trim-marker.nix"), []byte(`
    self: super:
    {
      # Start-marker: Nonsense
      ye = 1;
      # End-marker: Nonsense
      # Start-marker: Trim on install
      twlinst = import ./overlay_twlinst.nix {
        inherit self super;
      };
      # End-marker: Trim on install
    }`), 0644)
	ioutil.WriteFile(path.Join(dir, "trim-marker-unary.nix"), []byte(`
      self: super:
      {
        ye = 1;
        yeow = 2; # Line-marker: Trim on install
      }`), 0644)

	ioutil.WriteFile(path.Join(dir, "no-changes.nix"), []byte("abc\nd"), 0644)

	t.Run("trim on install multi", func(t *testing.T) {
		if err := (&Applyer{}).applyFile(path.Join(dir, "trim-marker.nix"), 0); err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadFile(path.Join(dir, "trim-marker.nix"))
		want := []byte(`
    self: super:
    {
      # Start-marker: Nonsense
      ye = 1;
      # End-marker: Nonsense
    }
`)

		if !bytes.Equal(b, want) {
			t.Errorf("Unexpected content:\n%s", string(b))
		}
	})

	t.Run("trim on install unary", func(t *testing.T) {
		if err := (&Applyer{}).applyFile(path.Join(dir, "trim-marker-unary.nix"), 0); err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadFile(path.Join(dir, "trim-marker-unary.nix"))
		want := []byte(`
      self: super:
      {
        ye = 1;
      }
`)

		if !bytes.Equal(b, want) {
			t.Errorf("Unexpected content:\n%s", string(b))
		}
	})

	t.Run("no changes", func(t *testing.T) {
		if err := (&Applyer{}).applyFile(path.Join(dir, "no-changes.nix"), 0); err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadFile(path.Join(dir, "no-changes.nix"))
		want := []byte("abc\nd")
		if !bytes.Equal(b, want) {
			t.Errorf("Unexpected content:\n%s", string(b))
		}
	})

}
