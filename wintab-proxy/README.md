# PPAP WinTab Proxy

This directory contains an experimental `Wintab32.dll` shim.

The proxy is intended to be copied next to a target Windows application executable so the app loads this DLL instead of the system/vendor WinTab DLL. It exposes a small WinTab-compatible surface and reads the latest PPAP packet from a named file mapping created by the PPAP server.

Current state:

- Exports common `Wintab32.dll` functions used by tablet-aware apps.
- Advertises one pressure-capable tablet device.
- Reads packets from `Local\\PPAPWinTabPacketV1`.
- Returns a fixed packet layout: `PK_X | PK_Y | PK_BUTTONS | PK_NORMAL_PRESSURE`.

Limitations:

- This is an experimental app-local proxy, not a system tablet driver.
- Some apps may bypass app-local DLL loading or require additional WinTab exports.
- A production implementation should either proxy to a real `Wintab32.dll` for unsupported calls or move to a virtual tablet driver.
