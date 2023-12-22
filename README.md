
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
 - [ ] Search
 - [ ] Landing