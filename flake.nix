{
  description = "mdreflow: semantic line breaks for Markdown and MDX";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        mdreflow = pkgs.buildGoModule {
          pname = "mdreflow";
          version = "0.3.0-fork.1";
          src = self;
          vendorHash = "sha256-8utHKF28vko8pwPlBgouRqhtyQPqtMNYqHOoMYXdSrg=";
          subPackages = [ "cmd/mdreflow" ];
          ldflags = [
            "-s"
            "-w"
            "-X main.version=v0.3.0-fork.1"
            "-X main.commit=${self.shortRev or "dirty"}"
            "-X main.date=${self.lastModifiedDate}"
          ];
          meta = {
            description = "Reformat Markdown and MDX to semantic line breaks";
            homepage = "https://github.com/jbeda/mdreflow";
            license = pkgs.lib.licenses.asl20;
            mainProgram = "mdreflow";
          };
        };
        default = mdreflow;
      });

      checks = self.packages;
    };
}
