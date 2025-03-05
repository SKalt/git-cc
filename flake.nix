{
  description = "A git extension to help write conventional commits.";
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, flake-utils, nixpkgs, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = (import nixpkgs) {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
        version = (builtins.fromJSON (builtins.readFile ./package.json)).version;
        rev = if (self ? rev) then self.rev else "dirty";

      in
      {
        packages.default = pkgs.callPackage ./. { inherit version; inherit rev; };
        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go # 1.23.x
          ];
          buildInputs = with pkgs; [
            nixpkgs-fmt
            nil
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
