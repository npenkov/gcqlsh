{
  description = "Cassandra command line shell written in Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = builtins.readFile ./VERSION;
        trimmedVersion = builtins.replaceStrings [ "\n" ] [ "" ] version;
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "gcqlsh";
            version = trimmedVersion;

            src = ./.;

            vendorHash = "sha256-LBD3E+IvOALSrSF5RqYE6pW/QHJJ9TfSBYnWKGU3aSk=";

            # Main package path
            subPackages = [ "cmd/gcqlsh" ];

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${trimmedVersion}"
            ];

            meta = with pkgs.lib; {
              description = "Cassandra command line shell written in Go";
              homepage = "https://github.com/npenkov/gcqlsh";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "gcqlsh";
            };
          };

          gcqlsh = self.packages.${system}.default;
        };

        apps = {
          default = flake-utils.lib.mkApp {
            drv = self.packages.${system}.default;
          };

          gcqlsh = self.apps.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            goreleaser
          ];

          shellHook = ''
            echo "gcqlsh development environment"
            echo "Go version: $(go version)"
          '';
        };
      });
}
