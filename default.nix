{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  ),
  version,
  rev,
}:

pkgs.buildGoApplication {
  pname = "git-cc";
  version = version + "+" + rev;
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  ldflags = ["-X" "main.version=${version}"];
}
