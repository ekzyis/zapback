let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/nixos-24.11";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShell {
  packages = with pkgs; [
    go
    tailwindcss
    gnumake
    inotify-tools
    figlet
    postgresql_17
  ];
  shellHook = ''
    # install templ if not already installed
    command -v templ > /dev/null || go install github.com/a-h/templ/cmd/templ@latest
  '';
}