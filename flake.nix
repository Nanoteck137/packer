{
  description = "Devshell and Building packer";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
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

          vendorHash = "sha256-e1IgVo2mAFk5BG9L2MIOKCA+5F3tJuG/p97kVQ50pQY=";

          nativeBuildInputs = [ pkgs.makeWrapper ];

          postFixup = ''
            wrapProgram $out/bin/packer --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.imagemagick ]}
          '';
        };
      in
      {
        packages.default = module;
        packages.packer = module;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            imagemagick
          ];
        };
      }
    );
}
