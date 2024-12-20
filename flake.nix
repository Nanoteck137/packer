{
  description = "Devshell and Building packer";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";

    devtools.url     = "github:nanoteck137/devtools";
    devtools.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, devtools, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        module = pkgs.buildGoModule {
          pname = "packer";
          version = self.shortRev or "dirty";
          src = ./.;

          vendorHash = "sha256-BJNahy9Gfxyc0/u0eCGMxo9jl2YVOrY/sJgQWENaji0=";

          nativeBuildInputs = [ pkgs.makeWrapper ];

          postFixup = ''
            wrapProgram $out/bin/packer --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.imagemagick ]}
          '';
        };

        tools = devtools.packages.${system};
      in
      {
        packages.default = module;
        packages.packer = module;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            imagemagick

            tools.publishVersion
          ];
        };
      }
    );
}
