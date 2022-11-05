{
  description = "A very basic flake";
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = (import nixpkgs) {
          inherit system;
        };
      in
      rec {
        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go # 1.19.x
          ];
          buildInputs = with pkgs; [
            nixpkgs-fmt
            rnix-lsp
            gopls
            gotools
            nodejs
            nodePackages.pnpm
            goreleaser
            ttyd
            ffmpeg
            vhs
            bashInteractive
          ];
        };
      }
    );
}
