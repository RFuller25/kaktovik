See also: [Kaktovik WearOS Watch](https://github.com/RFuller25/kaktovik-watch)

[https://rhysfuller.com/kaktovik/](https://rhysfuller.com/kaktovik/)

The kaktovik time system displays the current time using Inupiaq numerals, showing the day split into 20 parts, with those parts split into 20 parts, with those parts split into 20 parts, and finally those parts split into 20 parts. This gives an idea of the day as a fraction instead of arbitrarily going from 1-12.

## NixOS Integration

### Option 1: NixOS module (system-wide)

Add to your `flake.nix` inputs:

```nix
inputs = {
  kaktovik.url = "github:RFuller25/kaktovik";
};
```

Pass `inputs` through to your NixOS config and add the module + enable it:

```nix
# In your NixOS configuration (e.g. configuration.nix or a module it imports)
{ inputs, ... }:
{
  imports = [ inputs.kaktovik.nixosModules.default ];

  programs.kaktovik = {
    enable = true;
    # enableNotifications = true;  # default — installs libnotify for desktop alerts
  };
}
```

This installs `kaktovik` into `environment.systemPackages`.

### Option 2: Home Manager module (per-user)

```nix
# In your Home Manager configuration
{ inputs, ... }:
{
  imports = [ inputs.kaktovik.homeManagerModules.default ];

  programs.kaktovik.enable = true;
}
```

### Option 3: Direct package (no module)

If you just want the package without the module abstraction:

```nix
{ inputs, pkgs, ... }:
{
  environment.systemPackages = [
    inputs.kaktovik.packages.${pkgs.system}.default
  ];
}
```

### Running directly (no install)

```sh
nix run github:RFuller25/kaktovik
```

## Usage

```
kaktovik              # full TUI (clock, converter, timer, stopwatch, alarm)
kaktovik now          # print current Kaktovik time to stdout
kaktovik timer        # countdown timer
kaktovik stopwatch    # stopwatch
kaktovik alarm        # set an alarm
kaktovik convert      # time converter
```
