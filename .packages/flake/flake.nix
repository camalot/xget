{
  description = "A flake targeting multiple architectures";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixos-unstable/nixexprs.tar.zst";
  };
  
  outputs = { self, nixpkgs }: let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: builtins.listToAttrs (map (system: {
        name = system;
        value = f system;
      }) systems);
    in {
      packages = forAllSystems (system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        hello = pkgs.hello;
        default = pkgs.hello;
      });
    };
}