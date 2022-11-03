{
  description = "A very basic flake";
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      # packages.x86_64-linux.hello = nixpkgs.legacyPackages.x86_64-linux.hello;
      # packages.x86_64-linux.default = self.packages.x86_64-linux.hello;
      let
        pkgs = (import nixpkgs) {
          inherit system;
        };
      in
      rec {
        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go # 1.19
          ];
          buildInputs = with pkgs; [
            nixpkgs-fmt
            rnix-lsp
            gopls
            nodejs
            nodePackages.pnpm
            goreleaser
            ttyd
            ffmpeg
            vhs
          ];
        };
      }
    );
}
