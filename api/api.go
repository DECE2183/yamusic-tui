package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	_RESPONSE_TIMEOUT   = 2500 * time.Millisecond
	_TRACK_READ_TIMEOUT = 1500 * time.Millisecond
	_TIMESTAMP_FORMAT   = "2006-01-02T15:04:05.999Z"
)

var mTLSConfig = &tls.Config{
	CipherSuites: []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	},
	MinVersion: tls.VersionTLS12,
	// MaxVersion: tls.VersionTLS12,
}

var httpClient = http.Client{Transport: &http.Transport{
	TLSClientConfig:       mTLSConfig,
	ResponseHeaderTimeout: _RESPONSE_TIMEOUT,
}}

func nowTimestamp() string {
	return time.Now().Format(_TIMESTAMP_FORMAT)
}

func proccessRequest[RetT any](req *http.Request) (result RetT, invInfo InvocInfo, err error) {
	req.Header.Add("x-Yandex-Music-Client", "YandexMusicAndroid/24024312")
	req.Header.Add("User-Agent", "okhttp/4.12.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var respBody struct {
			InvocationInfo InvocInfo `json:"invocationInfo"`
			Result         RetT      `json:"result"`
		}

		dec := json.NewDecoder(resp.Body)
		dec.Decode(&respBody)

		invInfo = respBody.InvocationInfo
		result = respBody.Result
	case http.StatusBadRequest:
		var respBody struct {
			InvocationInfo InvocInfo       `json:"invocationInfo"`
			Error          BadRequestError `json:"error"`
		}

		dec := json.NewDecoder(resp.Body)
		dec.Decode(&respBody)

		invInfo = respBody.InvocationInfo
		err = respBody.Error
	case http.StatusUnauthorized:
		var respBody UnauthorizedError
		dec := json.NewDecoder(resp.Body)
		dec.Decode(&respBody)
		err = respBody
		invInfo.ReqId = respBody.RequestId
	default:
		err = fmt.Errorf("unhandled status %s", resp.Status)
	}

	return
}

func getRequest[RetT any](token, reqPath string, params url.Values) (result RetT, invInfo InvocInfo, err error) {
	reqUrl, err := url.JoinPath(YaMusicServerURL, reqPath)
	if err != nil {
		return
	}
	if params != nil {
		reqUrl += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "OAuth "+token)

	return proccessRequest[RetT](req)
}

func postRequest[RetT any](token, reqPath string, params url.Values) (result RetT, invInfo InvocInfo, err error) {
	reqUrl, err := url.JoinPath(YaMusicServerURL, reqPath)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "OAuth "+token)

	return proccessRequest[RetT](req)
}

func postRequestJson[RetT any](token, reqPath string, params url.Values, body any) (result RetT, invInfo InvocInfo, err error) {
	reqUrl, err := url.JoinPath(YaMusicServerURL, reqPath)
	if err != nil {
		return
	}
	if params != nil {
		reqUrl += "?" + params.Encode()
	}
	bodyData, err := json.Marshal(body)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewReader(bodyData))
	if err != nil {
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+token)

	return proccessRequest[RetT](req)
}

func downloadRequest(token, reqUrl, mimeType string) (body io.ReadCloser, contentLen int64, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		cancel()
		return
	}

	req.Header.Set("accept", mimeType)
	req.Header.Set("Authorization", "OAuth "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		cancel()
		return
	}

	if resp.StatusCode == 200 {
		body = NewTimeLimitedReader(resp.Body, ctx, cancel, _TRACK_READ_TIMEOUT)
		contentLen = resp.ContentLength
	} else {
		err = fmt.Errorf("error code %d", resp.StatusCode)
		resp.Body.Close()
		cancel()
	}

	return
}

func createTrackUrl(info fullDownloadInfo, codec string) string {
	trackUrl := "XGRlBW9FXlekgbPrRHuSiA" + info.Path[1:] + info.S
	hashSum := md5.Sum([]byte(trackUrl))
	hashedUrl := hex.EncodeToString(hashSum[:])
	return "https://" + info.Host + "/get-" + codec + "/" + hashedUrl + "/" + info.Ts + info.Path
}

// Deprecated: doesn't work in most cases
func Token(username, password string) (token string, err error) {
	params := url.Values{
		"grant_type":    {"password"},
		"client_id":     {yaOauthClientID},
		"client_secret": {yaOauthClientSecret},
		"username":      {username},
		"password":      {password},
	}

	servPath, err := url.JoinPath(yaOauthServerURL, "token")
	if err != nil {
		return
	}
	resp, err := http.Post(servPath, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody := map[string]string{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&respBody)

	errDesc, ok := respBody["error_description"]
	if ok {
		err = fmt.Errorf(errDesc)
		return
	}

	token, ok = respBody["access_token"]
	if !ok {
		err = fmt.Errorf("unknown response format")
		return
	}

	return
}

func ShareTrackLink(track *Track) string {
	if len(track.Albums) == 0 {
		return ""
	}
	return fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
}

func TrackCoverLink(track *Track, size int) string {
	if len(track.CoverUri) < 2 {
		return ""
	}
	return fmt.Sprintf("https://%s%dx%d", track.CoverUri[:len(track.CoverUri)-2], size, size)
}

func DownloadTrackCover(dst io.Writer, track *Track, size int) (string, error) {
	url := TrackCoverLink(track, size)
	if len(url) == 0 {
		return "", errors.New("cover not presented")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	_, err = io.Copy(dst, resp.Body)
	return resp.Header.Get("Content-Type"), err
}

func NewClient(name, token string) (*YaMusicClient, error) {
	client := &YaMusicClient{
		name:  name,
		token: token,
	}

	clientStatus, _, err := getRequest[UserStatus](token, "account/status", nil)
	if err != nil {
		return nil, err
	}

	client.userid = clientStatus.Account.Uid
	client.sessionid = nowTimestamp()

	return client, nil
}

func (client *YaMusicClient) Tracks(trackIds []string) (tracks []Track, err error) {
	tracks, _, err = postRequest[[]Track](client.token, "/tracks", url.Values{"track-ids": trackIds, "with-positions": {"false"}})
	return
}

func (client *YaMusicClient) CreatePlaylist(name string, public bool) (playlist Playlist, err error) {
	var visibility string
	if public {
		visibility = "public"
	} else {
		visibility = "private"
	}
	playlist, _, err = postRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/create", client.userid), url.Values{
		"title":      {name},
		"visibility": {visibility},
	})
	return
}

func (client *YaMusicClient) RenamePlaylist(kind uint64, newName string) (playlist Playlist, err error) {
	playlist, _, err = postRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/%d/name", client.userid, kind), url.Values{
		"value": {newName},
	})
	return
}

func (client *YaMusicClient) RemovePlaylist(kind uint64) error {
	_, _, err := postRequest[string](client.token, fmt.Sprintf("/users/%d/playlists/%d/delete", client.userid, kind), nil)
	return err
}

func (client *YaMusicClient) AddToPlaylist(kind uint64, revision, pos int, trackId string) (playlist Playlist, err error) {
	playlist, _, err = postRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/%d/change-relative", client.userid, kind), url.Values{
		"diff":     {fmt.Sprintf(`{"diff":{"op":"insert","at":%d,"tracks":[{"id":"%s"}]}}`, pos, trackId)},
		"revision": {fmt.Sprint(revision)},
	})
	return playlist, err
}

func (client *YaMusicClient) RemoveFromPlaylist(kind uint64, revision, pos int) (playlist Playlist, err error) {
	playlist, _, err = postRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/%d/change-relative", client.userid, kind), url.Values{
		"diff":     {fmt.Sprintf(`{"diff":{"op":"delete","from":%d,"to":%d}}`, pos, pos+1)},
		"revision": {fmt.Sprint(revision)},
	})
	return playlist, err
}

func (client *YaMusicClient) ListPlaylists() (playlists []Playlist, err error) {
	playlists, _, err = getRequest[[]Playlist](client.token, fmt.Sprintf("/users/%d/playlists/list", client.userid), nil)
	return
}

func (client *YaMusicClient) Playlist(kind uint64) (playlist Playlist, err error) {
	playlist, _, err = getRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/%d", client.userid, kind), nil)
	return
}

func (client *YaMusicClient) PlaylistTracks(kind uint64, userId uint64, mixed bool) (tracks []Track, err error) {
	params := url.Values{
		"kinds":       {fmt.Sprint(kind)},
		"mixed":       {fmt.Sprint(mixed)},
		"rich-tracks": {"true"},
	}

	playlists, _, err := getRequest[[]Playlist](client.token, fmt.Sprintf("/users/%d/playlists", userId), params)
	if err != nil {
		return
	}

	if len(playlists) != 1 {
		err = fmt.Errorf("wrong playlists count")
		return
	}

	tracks = make([]Track, 0, playlists[0].TrackCount)
	for i := 0; i < playlists[0].TrackCount; i++ {
		tracks = append(tracks, playlists[0].Tracks[i].Track)
	}

	return
}

func (client *YaMusicClient) Stations(language string) (stations []StationDesc, err error) {
	stations, _, err = getRequest[[]StationDesc](client.token, "/rotor/stations/list", url.Values{
		"language": {language},
	})
	return
}

func (client *YaMusicClient) StationTracks(id StationId, lastTrack *Track) (tracks StationTracks, err error) {
	params := url.Values{
		"settings2": {"true"},
	}
	if lastTrack != nil {
		params.Add("queue", fmt.Sprint(lastTrack.Id))
	}
	tracks, _, err = getRequest[StationTracks](client.token, fmt.Sprintf("/rotor/station/%s:%s/tracks", id.Type, id.Tag), nil)
	return
}

func (client *YaMusicClient) StationFeedback(feedType string, stationId StationId, batchId, trackId string, playedSeconds int) (err error) {
	queryParams := url.Values{}
	if len(batchId) > 0 {
		queryParams.Add("batch-id", batchId)
	}

	body := map[string]interface{}{
		"type":               feedType,
		"timestamp":          nowTimestamp(),
		"from":               client.name,
		"trackId":            trackId,
		"totalPlayedSeconds": playedSeconds,
	}
	_, _, err = postRequestJson[interface{}](client.token,
		fmt.Sprintf("/rotor/station/%s:%s/feedback", stationId.Type, stationId.Tag),
		queryParams,
		body,
	)
	return
}

func (client *YaMusicClient) PlayTrack(track *Track, fromCache bool) (err error) {
	queryParams := url.Values{
		"uid":                  {fmt.Sprint(client.userid)},
		"from":                 {client.name},
		"play-id":              {client.sessionid},
		"track-id":             {track.Id},
		"from-cache":           {fmt.Sprint(fromCache)},
		"track-length-seconds": {fmt.Sprint(track.DurationMs + 1000)},
		"total-played-seconds": {fmt.Sprint(track.DurationMs + 1000)},
		"timestamp":            {nowTimestamp()},
	}
	_, _, err = postRequest[interface{}](client.token, "/play-audio", queryParams)
	return
}

func (client *YaMusicClient) LikedTracks() (tracks []LikeTrackInfo, err error) {
	desc, _, err := getRequest[LikesDesc](client.token, fmt.Sprintf("/users/%d/likes/tracks", client.userid), nil)
	if err != nil {
		return
	}
	tracks = desc.Library.Tracks
	return
}

func (client *YaMusicClient) LikeTrack(trackId string) (err error) {
	_, _, err = postRequest[interface{}](client.token, fmt.Sprintf("/users/%d/likes/tracks/add-multiple", client.userid), url.Values{"track-ids": {trackId}})
	return
}

func (client *YaMusicClient) UnlikeTrack(trackId string) (err error) {
	_, _, err = postRequest[interface{}](client.token, fmt.Sprintf("/users/%d/likes/tracks/remove", client.userid), url.Values{"track-ids": {trackId}})
	return
}

func (client *YaMusicClient) TrackDownloadInfo(trackId string) (dowInfos []TrackDownloadInfo, err error) {
	dowInfos, _, err = getRequest[[]TrackDownloadInfo](client.token, fmt.Sprintf("/tracks/%s/download-info", trackId), nil)
	return
}

func (client *YaMusicClient) DownloadTrack(dowInfo TrackDownloadInfo) (track io.ReadCloser, fileSize int64, err error) {
	fullInfoBody, _, err := downloadRequest(client.token, dowInfo.DownloadInfoUrl+"&format=json", "application/json")
	if err != nil {
		return
	}

	var info fullDownloadInfo
	dec := json.NewDecoder(fullInfoBody)
	err = dec.Decode(&info)
	fullInfoBody.Close()
	if err != nil {
		return
	}

	var mimeType string
	switch dowInfo.Codec {
	case "aac":
		mimeType = "audio/aac"
	case "mp3":
		mimeType = "audio/mpeg"
	default:
		err = fmt.Errorf("unknown codec type '%s'", dowInfo.Codec)
		return
	}

	trackUrl := createTrackUrl(info, dowInfo.Codec)
	trackReader, fileSize, err := downloadRequest(client.token, trackUrl, mimeType)
	track = trackReader
	return
}

func (client *YaMusicClient) ArtistTracks(artistId uint64, page, pageSize int) (tracks ArtistTracks, err error) {
	tracks, _, err = getRequest[ArtistTracks](client.token,
		fmt.Sprintf("/artists/%d/tracks", artistId),
		url.Values{"page": {fmt.Sprint(page)}, "page-size": {fmt.Sprint(pageSize)}},
	)
	return
}

func (client *YaMusicClient) ArtistPopularTracks(artistId uint64) (tracks ArtistTracks, err error) {
	tracks, _, err = getRequest[ArtistTracks](client.token, fmt.Sprintf("/artists/%d/track-ids-by-rating", artistId), nil)
	return
}

func (client *YaMusicClient) Album(albumId uint64, withTracks bool) (album Album, err error) {
	path := fmt.Sprintf("/albums/%d", albumId)
	if withTracks {
		path += "/with-tracks"
	}
	album, _, err = getRequest[Album](client.token, path, nil)
	return
}

func (client *YaMusicClient) Search(request string, searchType SearchType) (results SearchResult, err error) {
	results, _, err = getRequest[SearchResult](client.token, "/search", url.Values{"text": {request}, "page": {"0"}, "type": {string(searchType)}})
	for i := range results.Tracks.Results {
		results.Tracks.Results[i].Id = results.Tracks.Results[i].RealId
	}
	return
}

func (client *YaMusicClient) SearchSuggest(part string) (suggestions SearchSuggest, err error) {
	suggestions, _, err = getRequest[SearchSuggest](client.token, "/search/suggest", url.Values{"part": {part}})
	return
}

func (client *YaMusicClient) TrackLyricsRequest(trackId string) (LRCLyrics []LyricPair, err error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	// scary algorithm to sign the request (required for lyrics)
	message := trackId + timestamp
	h := hmac.New(sha256.New, []byte("p93jhgh689SBReK6ghtw62"))
	h.Write([]byte(message))
	hmacSign := h.Sum(nil)
	sign := base64.StdEncoding.EncodeToString(hmacSign)
	lyrics, _, err := getRequest[TrackLyrics](client.token, fmt.Sprintf("/tracks/%s/lyrics", trackId), url.Values{"sign": {sign}, "timeStamp": {timestamp}, "format": {"LRC"}})
	if err != nil {
		return []LyricPair{}, err
	}
	LRCLyricsResponse, err := http.Get(lyrics.DownloadUrl)
	if err != nil {
		return []LyricPair{}, err
	}
	data, err := io.ReadAll(LRCLyricsResponse.Body)
	if err != nil {
		return []LyricPair{}, err
	}
	LRCLyrics = parseLRCText(string(data))
	return
}

func parseLRCText(lrcContent string) []LyricPair {
	var lyrics []LyricPair
	lines := strings.Split(lrcContent, "\n")

	for _, line := range lines {
		if !strings.Contains(line, "[") || strings.HasPrefix(line, "[ti:") {
			continue
		}

		parts := strings.SplitN(line, "]", 2)
		if len(parts) != 2 {
			continue
		}

		timeStr := strings.Trim(parts[0], "[]")
		timeParts := strings.Split(timeStr, ":")
		if len(timeParts) != 2 {
			continue
		}

		minutes, _ := strconv.Atoi(timeParts[0])
		secondsParts := strings.Split(timeParts[1], ".")
		seconds, _ := strconv.Atoi(secondsParts[0])
		millis := 0
		if len(secondsParts) > 1 {
			millis, _ = strconv.Atoi(secondsParts[1])
		}

		totalMs := minutes*60*1000 + seconds*1000 + millis
		lyrics = append(lyrics, LyricPair{totalMs, strings.TrimSpace(parts[1])})
	}

	return lyrics
}
