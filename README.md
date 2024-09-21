
# yamusic-tui

[![GitHub License](https://img.shields.io/github/license/dece2183/yamusic-tui)](https://github.com/DECE2183/yamusic-tui/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dece2183/yamusic-tui)](https://goreportcard.com/report/github.com/dece2183/yamusic-tui)
[![Release](https://img.shields.io/github/v/release/dece2183/yamusic-tui)](https://github.com/dece2183/yamusic-tui/releases)

An unofficial Yandex Music terminal client.<br>
Based on [yandex-music-open-api](https://github.com/acherkashin/yandex-music-open-api).

![screenshot](.assets/screenshot.png)

### Requirements

To use this client, you should have a valid Yandex Music account and an access token.<br>
The easiest way to get a token is to use this
[browser extension](https://github.com/MarshalX/yandex-music-token/tree/main/browser-extension).

### Implemented features

 - [x] Player control
    - [x] Play/pause
    - [x] Switch track
    - [x] Play progress
    - [x] Rewind
    - [x] Like/unlike
    - [x] Share
 - [ ] Radio
    - [x] My wave
    - [ ] Radio configuration
 - [ ] Likes
    - [x] Liked tracks
    - [ ] Liked playlists
    - [ ] Liked artists
    - [ ] Liked albums
 - [ ] Playlists
    - [x] Display user playlists
    - [x] Play from playlist
    - [ ] Add/remove track to playlist
    - [ ] Create/remove playlist
    - [ ] Rename playlist
 - [ ] Caching
 - [x] Search
 - [ ] Landing

## Installation

If you have Go installed on your PC:

```bash
go install github.com/dece2183/yamusic-tui@latest
```

## Configuration

The configuration file is located at `~/.config/yamusic-tui/config.yaml`.

This is the default configuration which is automatically created after the first login:

```yaml
token: <your yandex music token>
buffer-size-ms: 80
rewind-duration-s: 5
volume: 0.5
volume-step: 0.05
controls:
   quit: ctrl+q,ctrl+c
   apply: enter
   cancel: esc
   cursor-up: up
   cursor-down: down
   playlists-up: ctrl+up
   playlists-down: ctrl+down
   tracks-like: l
   tracks-add-to-playlist: a
   tracks-share: ctrl+s
   tracks-shuffle: ctrl+x
   tracks-search: ctrl+f
   player-pause: space
   player-next: right
   player-previous: left
   player-rewind-forward: ctrl+right
   player-rewind-backward: ctrl+left
   player-like: L
   player-vol-up: +,=
   player-vol-donw: '-'
```

You can list multiple keys for the same control, separated by commas.

Increase the `buffer-size-ms` if you have glitches or statters.
