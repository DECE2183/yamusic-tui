package api

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	MaxVersion: tls.VersionTLS12,
}

var client = http.Client{Transport: &http.Transport{TLSClientConfig: mTLSConfig}}

func (e ResultError) Error() string {
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
}

func nowTimestamp() string {
	nowTime := time.Now()
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d",
		nowTime.Year(), nowTime.Month(), nowTime.Day(),
		nowTime.Hour(), nowTime.Minute(), nowTime.Second(),
	)
}

func proccessRequest[RetT any](req *http.Request) (result RetT, invInfo InvocInfo, err error) {
	req.Header.Add("x-Yandex-Music-Client", "YandexMusicAndroid/24024312")
	req.Header.Add("User-Agent", "okhttp/4.12.0")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var respBody struct {
			InvocationInfo InvocInfo `json:"invocationInfo"`
			Result         RetT      `json:"result"`
		}

		dec := json.NewDecoder(resp.Body)
		dec.Decode(&respBody)

		invInfo = respBody.InvocationInfo
		result = respBody.Result
	} else {
		var respBody struct {
			InvocationInfo InvocInfo   `json:"invocationInfo"`
			Error          ResultError `json:"error"`
		}

		dec := json.NewDecoder(resp.Body)
		dec.Decode(&respBody)

		invInfo = respBody.InvocationInfo
		err = respBody.Error
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
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return
	}

	req.Header.Set("accept", mimeType)
	req.Header.Set("Authorization", "OAuth "+token)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		body = resp.Body
		contentLen = resp.ContentLength
	} else {
		err = fmt.Errorf("error code %d", resp.StatusCode)
		resp.Body.Close()
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

func NewClient(token string) (client *YaMusicClient, err error) {
	client = &YaMusicClient{
		token: token,
	}

	clientStatus, _, err := getRequest[UserStatus](token, "account/status", nil)
	client.userid = clientStatus.Account.Uid

	return
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

func (client *YaMusicClient) AddToPlaylist(kind uint64, revision, pos int, trackId string) (playlist Playlist, err error) {
	playlist, _, err = postRequest[Playlist](client.token, fmt.Sprintf("/users/%d/playlists/%d/change-relative", client.userid, kind), url.Values{
		"diff":     {fmt.Sprintf(`{"diff":{"op":"insert","at":%d,"tracks":[{"id":"%s"}]}}`, pos, trackId)},
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

func (client *YaMusicClient) PlaylistTracks(kind uint64, mixed bool) (tracks []Track, err error) {
	params := url.Values{
		"kinds":       {fmt.Sprint(kind)},
		"mixed":       {fmt.Sprint(mixed)},
		"rich-tracks": {"true"},
	}

	playlists, _, err := getRequest[[]Playlist](client.token, fmt.Sprintf("/users/%d/playlists", client.userid), params)
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
		"from":               "yamusic-tui",
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
		"from":                 {"yamusic-tui"},
		"uid":                  {fmt.Sprint(client.userid)},
		"timestamp":            {nowTimestamp()},
		"track-id":             {track.Id},
		"from-cache":           {fmt.Sprint(fromCache)},
		"track-length-seconds": {fmt.Sprint(track.DurationMs + 1000)},
		"total-played-seconds": {fmt.Sprint(track.DurationMs + 1000)},
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

func (client *YaMusicClient) DownloadTrack(dowInfo TrackDownloadInfo) (track *HttpReadSeeker, fileSize int64, err error) {
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
	track = newReadSeaker(trackReader, fileSize)
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
