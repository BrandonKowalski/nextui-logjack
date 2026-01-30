<!-- trunk-ignore-all(markdownlint/MD033) -->
<!-- trunk-ignore(markdownlint/MD041) -->
<div align="center">

  <h1>Log Jack</h1>
  <h3>
    A wireless log fetching tool for <a href="https://nextui.loveretro.games">NextUI</a>.
  </h3>

<br>

[![license-badge-img]][license-badge]
[![release-badge-img]][release-badge]
[![downloads-badge-img]][downloads-badge]

</div>

---

## Features

- Browse files on your device from any web browser
- Download logs, saves, or any configured directory
- QR code for easy mobile access
- Configurable directories via INI file
- Environment variable support in paths

---

## Configuration

Edit `logjack.ini` inside the pak to configure directories:

```ini
[directories]
Logs = /mnt/.userdata/{PLATFORM}/logs
Saves = /mnt/SDCARD/Saves
Roms = /mnt/SDCARD/Roms
```

Environment variables can be used with `{VAR}` syntax (e.g., `{PLATFORM}` expands to `tg5040` or `tg5050`).

---

## Usage

1. Launch Log Jack from the Tools menu
2. Scan the QR code with your phone or visit the displayed URL
3. Browse and download files

Press **B** to quit.

---

## Need Help? Found a Bug?

Please [create an issue](https://github.com/BrandonKowalski/nextui-logjack/issues/new).

---

<!-- Badges -->

[license-badge-img]: https://img.shields.io/github/license/BrandonKowalski/nextui-logjack?style=for-the-badge&color=00d4aa

[license-badge]: LICENSE

[release-badge-img]: https://img.shields.io/github/v/release/BrandonKowalski/nextui-logjack?sort=semver&style=for-the-badge&color=00d4aa

[release-badge]: https://github.com/BrandonKowalski/nextui-logjack/releases

[downloads-badge-img]: https://img.shields.io/github/downloads/BrandonKowalski/nextui-logjack/total?style=for-the-badge&color=00d4aa

[downloads-badge]: https://github.com/BrandonKowalski/nextui-logjack/releases
