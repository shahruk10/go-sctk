{ pkgs ? import (fetchTarball
  "https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz") { } }:
let
  stdenv = pkgs.pkgsMusl.stdenv;
  mkShell = pkgs.mkShell.override { inherit stdenv; };
in mkShell {
  buildInputs = with pkgs; [ git go_1_18 perl cacert ];
}
