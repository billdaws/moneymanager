{
  description = "Go-based finance management application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go
            gopls          # Language server
            gotools        # goimports, etc.
            golangci-lint  # Linting

            # Docker and compose
            docker
            docker-compose

            # SQLite tools
            sqlite
            sqlite-analyzer

            # Utilities
            curl
            jq

            # Code formatting
            gofumpt        # Stricter gofmt
          ];

          shellHook = ''
            echo "Finance Manager Dev Environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go run cmd/server/main.go  - Run server"
            echo "  golangci-lint run          - Run linters"
            echo "  docker-compose up -d       - Start Kreuzberg"
            echo "  docker-compose down        - Stop Kreuzberg"
            echo ""

            # Set up Go environment
            export CGO_ENABLED=1

            # Create necessary directories
            mkdir -p data logs uploads
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "moneymanager";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;  # Will update after first build
        };
      }
    );
}
