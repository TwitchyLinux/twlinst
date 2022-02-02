with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    gtk3 pkg-config
  ];
}
