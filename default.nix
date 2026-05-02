{ pkgs ? import <nixpkgs> {}, version ? "unstable" }:

let
  buildGoModule =
    if pkgs ? buildGo124Module then
      pkgs.buildGo124Module
    else
      pkgs.buildGoModule.override { go = pkgs.go_1_24; };
in
buildGoModule {
  pname = "x11-sleep-manager";
  inherit version;

  src = ./.;
  vendorHash = "sha256-Ac63bZlBvCrhS7b8mk7aJdApI8UGtJxnZG35L37roGY=";
  subPackages = [
    "cmd/x11-sleep-manager"
    "cmd/x11smctl"
  ];

  nativeBuildInputs = [ pkgs.makeWrapper ];

  postInstall = ''
    runtimePath=${pkgs.lib.makeBinPath [ pkgs.xorg.xset ]}
    wrapProgram "$out/bin/x11-sleep-manager" --prefix PATH : "$runtimePath"
    wrapProgram "$out/bin/x11smctl" --prefix PATH : "$runtimePath"
  '';

  meta = with pkgs.lib; {
    description = "X11/i3 daemon bridging logind inhibitors to X11 lock behavior";
    homepage = "https://github.com/jvalduvieco/x11-sleep-manager";
    license = licenses.mit;
    platforms = platforms.linux;
    mainProgram = "x11-sleep-manager";
  };
}
