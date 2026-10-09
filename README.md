# Giddy

A tiny self-hosted photo and video gallery for your LAN. Point it at a folder, open it in a browser.

- **Go backend** (~150 lines, stdlib only): scans the folder, serves files with Range support so video seeking works.
- **Single-file frontend**: no build step, no CDN, no JS dependencies. Works fully offline.
- **Caddy** in front for basic auth.
- Picks up new files automatically (polls every 5s; the disk scan is cached for 3s).

## Run

```sh
git clone <this-repo> giddy && cd giddy

# 1. Create your Caddyfile and set a password (the real one is gitignored)
cp Caddyfile.example Caddyfile
docker run --rm caddy:2-alpine caddy hash-password --plaintext 'your-password'
# paste the output over REPLACE_WITH_HASH in Caddyfile

# 2. Add some media and start it
cp -r ~/Pictures/holiday media/
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

## Config

| Env var     | Default  | Purpose              |
|-------------|----------|----------------------|
| `MEDIA_DIR` | `/media` | Folder to scan/serve |
| `ADDR`      | `:8080`  | Backend listen addr  |

## License

MIT, see [LICENSE](LICENSE).
