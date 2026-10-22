# Giddy

A tiny self-hosted photo and video gallery for your LAN. Point it at a folder, open it in a browser.

- **Go backend** (~160 lines, stdlib only): scans the folder, serves files with Range support so video seeking works.
- **Single-file frontend**: no build step, no CDN, no JS dependencies. Works fully offline.
- **Rich features**: Instant search, folder/album filtering, flexible sorting (Date, Name, Size, Type), density toggles, lightbox slideshow, mobile swipe gestures, and 1-click downloads.
- **Caddy** in front for basic auth (or optional passwordless mode).
- Picks up new files automatically (polls every 5s; the disk scan is cached for 3s).

## Run

```sh
git clone <this-repo> giddy && cd giddy

# 1. Setup environment & auth
cp .env.example .env

# Option A: Passwordless mode (no login prompt)
# Set PASSWORDLESS=true in .env

# Option B: Basic auth (default)
# Generate a password hash and replace REPLACE_WITH_HASH in Caddyfile:
docker run --rm caddy:2-alpine caddy hash-password --plaintext 'your-password'

# 2. Configure media folder
# Either drop files into ./media or set MEDIA_PATH in .env

# 3. Start it
docker compose up -d --build
```

Open `http://<host-ip>/` and log in as `admin` with the password you chose.

To change the password later, generate a new hash the same way, replace it in `Caddyfile`, then `docker compose restart caddy`.

## Notes

- Supported: jpg, jpeg, png, gif, webp, avif, mp4, webm, ogg, mov, mkv. Whether a video *plays* depends on the browser's codec support (mkv and some mov files won't).
- Hidden files and folders (`.DS_Store`, `._*.jpg`) are ignored; directory listings are disabled.
- The media folder is mounted read-only. The app can't modify your files.
- Newest files first (by modified time).
- Auth is over plain HTTP. On a trusted home LAN that's usually fine; for HTTPS, use a hostname in the Caddyfile with `tls internal` and trust Caddy's root CA on your devices.

## Shortcuts & Controls

- **`Left` / `Right`**: Previous / Next media in lightbox (or swipe left/right on touchscreens)
- **`Space`**: Start / Pause slideshow
- **`Esc`**: Close lightbox (browser Back button also closes it)
- **Grid density**: Toggle between ▦ Compact, ⊞ Standard, and ⊡ Large previews

## Config

Environment variables can be configured in a `.env` file (see `.env.example`):

| Env var        | Default   | Purpose                                    |
|----------------|-----------|--------------------------------------------|
| `MEDIA_PATH`   | `./media` | Host folder containing photos/videos       |
| `PORT`         | `80`      | Host HTTP port exposed by Caddy            |
| `PASSWORDLESS` | `false`   | Set to `true` to disable basic auth        |
| `MEDIA_DIR`    | `/media`  | Backend folder to scan/serve (in container)|
| `ADDR`         | `:8080`   | Backend listen addr (in container)         |

## License

MIT, see [LICENSE](LICENSE).
