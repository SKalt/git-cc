{
  description = "A very basic flake";
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "flake-utils";
    };

  };

  outputs = { self, flake-utils, nixpkgs, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = (import nixpkgs) {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
      in {
        packages.default = pkgs.callPackage ./. { };
        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go # 1.20.x
          ];
          buildInputs = with pkgs; [
            nixpkgs-fmt
            rnix-lsp
            gopls
            gotools
            libfaketime
            glibc
            git
            nodejs
            gomod2nix.packages.${system}.default
            coreutils-full
            toybox
            vim
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
