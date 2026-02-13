{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    golangci-lint
    docker
    docker-compose
    sqlite
    sqlite-analyzer
    curl
    jq
    gofumpt
  ];

  shellHook = ''
    export CGO_ENABLED=1
    mkdir -p data logs uploads
    echo "Finance Manager Dev Environment Ready"
    echo "Go version: $(go version)"
  '';
}
