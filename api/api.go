package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (e ResultError) Error() string {
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
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

	resp, err := http.DefaultClient.Do(req)
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

	resp, err := http.DefaultClient.Do(req)
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

func downloadRequest(token, reqUrl, mimeType string) (body io.ReadCloser, contentLen int64, err error) {
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return
	}

	req.Header.Set("accept", mimeType)
	req.Header.Set("Authorization", "OAuth "+token)

	resp, err := http.DefaultClient.Do(req)
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
	track, fileSize, err = downloadRequest(client.token, trackUrl, mimeType)
	return
}
