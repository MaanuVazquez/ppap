# PPAP MVP

PPAP is an MVP for using an iPad with Apple Pencil as a remote pressure-sensitive input surface for a Windows desktop.

## Scope

This first version targets:

- Windows 10/11 x64 server.
- iPad browser client.
- Local network only.
- One Windows desktop surface.
- WebSocket pen event transport.
- Windows Ink synthetic pen injection.
- Basic JPEG screen preview polling.

The initial screen preview is intentionally simple. The make-or-break feature is whether Windows Ink synthetic pen input is accepted by Krita and Photoshop in Windows Ink mode.

## Requirements

- Go 1.22 or newer.
- Node.js 20 or newer.
- Windows 10/11 x64 for real pen injection and screen capture.

## Run In Development

Terminal 1:

```sh
cd server
go run ./cmd/ppap-server -addr :4040
```

Terminal 2:

```sh
cd client
npm install
npm run dev -- --host 0.0.0.0
```

Open the Vite URL from the iPad. The Vite dev server proxies `/api` requests to the Go server on port `4040`.

## Build Client For Server Hosting

```sh
cd client
npm install
npm run build
```

Then run:

```sh
cd server
go run ./cmd/ppap-server -client-dir ../client/dist -addr :4040
```

Open the Windows machine LAN URL from the iPad, for example:

```txt
http://192.168.1.50:4040
```

## Create A Windows Release

Every push to `main` builds a standalone Windows x64 binary and creates a GitHub Release named after the commit SHA, for example `build-2493eeb`.

Push a version tag when you want a named version release:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds the client, embeds it into the Go server, cross-compiles a Windows x64 `.exe`, and uploads a `.zip` plus `SHA256SUMS`.

You can also run the `Release Windows x64` workflow manually from GitHub Actions with a release tag input.

## App Setup Notes

For Krita and Photoshop testing:

- Enable Windows Ink mode in the app/tablet settings.
- Select a brush configured to use pressure for size or opacity.
- Keep the Go server and target app in the same Windows desktop session.
- If the target app runs elevated, run the server elevated too.
- Use the `Pressure Test` button in the iPad client to inject a server-generated pressure ramp into the focused drawing app.
- The backend switch currently exposes Windows Ink and a WinTab scaffold. WinTab is reported as unavailable until PPAP has a virtual tablet driver or Wintab32 proxy, because WinTab does not provide a global user-mode injection API.
- Open `/api/diagnostics` on the server to confirm the active backend, foreground Windows window, and virtual desktop bounds while testing.

## Current Limitations

- Screen preview is JPEG polling, not low-latency WebRTC.
- Synthetic pen injection is Windows x64 only.
- Non-Windows builds use no-op stubs so the project can compile elsewhere.
- Multi-monitor calibration is not implemented yet.
