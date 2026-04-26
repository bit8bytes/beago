{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  outputs = {
    self,
    nixpkgs,
  }: let
    # Production server and Apple Silicon dev machines (aarch64-darwin).
    supportedSystems = [
      "x86_64-linux"
      "aarch64-darwin"
    ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
  in {
    # Local development shell with all required tools.
    # Use `nix develop` to open the dev shell.
    devShells = forAllSystems (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      default = pkgs.mkShell {
        packages = with pkgs; [
          git
          go
          go-task
        ];
      };
    });
  };
}
