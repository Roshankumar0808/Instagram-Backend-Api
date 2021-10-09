package goinsta

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Instagram struct {
	user string
	pass string

	dID string

	uuid string

	rankToken string

	token string

	pid string

	adid string

	challengeURL string

	Challenge *Challenge

	Profiles *Profiles

	Account *Account

	Search *Search

	Timeline *Timeline

	Activity *Activity

	Inbox *Inbox

	Feed *Feed

	Contacts *Contacts

	Locations *LocationInstance

	c *http.Client
}

func (inst *Instagram) SetHTTPClient(client *http.Client) {
	inst.c = client
}

func (inst *Instagram) SetHTTPTransport(transport http.RoundTripper) {
	inst.c.Transport = transport
}

func (inst *Instagram) SetDeviceID(id string) {
	inst.dID = id
}

func (inst *Instagram) SetUUID(uuid string) {
	inst.uuid = uuid
}

func (inst *Instagram) SetPhoneID(id string) {
	inst.pid = id
}

func (inst *Instagram) SetCookieJar(jar http.CookieJar) error {
	url, err := neturl.Parse(goInstaAPIUrl)
	if err != nil {
		return err
	}

	cookies := inst.c.Jar.Cookies(url)
	inst.c.Jar = jar
	inst.c.Jar.SetCookies(url, cookies)
	return nil
}

func New(username, password string) *Instagram {

	jar, _ := cookiejar.New(nil)
	inst := &Instagram{
		user: username,
		pass: password,
		dID: generateDeviceID(
			generateMD5Hash(username + password),
		),
		uuid: generateUUID(),
		pid:  generateUUID(),
		c: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Jar: jar,
		},
	}
	inst.init()

	return inst
}

func (inst *Instagram) init() {
	inst.Challenge = newChallenge(inst)
	inst.Profiles = newProfiles(inst)
	inst.Activity = newActivity(inst)
	inst.Timeline = newTimeline(inst)
	inst.Search = newSearch(inst)
	inst.Inbox = newInbox(inst)
	inst.Feed = newFeed(inst)
	inst.Contacts = newContacts(inst)
	inst.Locations = newLocation(inst)
}

func (inst *Instagram) SetProxy(url string, insecure bool) error {
	uri, err := neturl.Parse(url)
	if err == nil {
		inst.c.Transport = &http.Transport{
			Proxy: http.ProxyURL(uri),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		}
	}
	return err
}

func (inst *Instagram) UnsetProxy() {
	inst.c.Transport = nil
}

func (inst *Instagram) Save() error {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("home")
	}
	return inst.Export(filepath.Join(home, ".goinsta"))
}

func (inst *Instagram) Export(path string) error {
	url, err := neturl.Parse(goInstaAPIUrl)
	if err != nil {
		return err
	}

	config := ConfigFile{
		ID:        inst.Account.ID,
		User:      inst.user,
		DeviceID:  inst.dID,
		UUID:      inst.uuid,
		RankToken: inst.rankToken,
		Token:     inst.token,
		PhoneID:   inst.pid,
		Cookies:   inst.c.Jar.Cookies(url),
	}
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}

func Export(inst *Instagram, writer io.Writer) error {
	url, err := neturl.Parse(goInstaAPIUrl)
	if err != nil {
		return err
	}

	config := ConfigFile{
		ID:        inst.Account.ID,
		User:      inst.user,
		DeviceID:  inst.dID,
		UUID:      inst.uuid,
		RankToken: inst.rankToken,
		Token:     inst.token,
		PhoneID:   inst.pid,
		Cookies:   inst.c.Jar.Cookies(url),
	}
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	return err
}

func ImportReader(r io.Reader) (*Instagram, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	config := ConfigFile{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return ImportConfig(config)
}

func ImportConfig(config ConfigFile) (*Instagram, error) {
	url, err := neturl.Parse(goInstaAPIUrl)
	if err != nil {
		return nil, err
	}

	inst := &Instagram{
		user:      config.User,
		dID:       config.DeviceID,
		uuid:      config.UUID,
		rankToken: config.RankToken,
		token:     config.Token,
		pid:       config.PhoneID,
		c: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
	inst.c.Jar, err = cookiejar.New(nil)
	if err != nil {
		return inst, err
	}
	inst.c.Jar.SetCookies(url, config.Cookies)

	inst.init()
	inst.Account = &Account{inst: inst, ID: config.ID}
	inst.Account.Sync()

	return inst, nil
}

func Import(path string) (*Instagram, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ImportReader(f)
}

func (inst *Instagram) readMsisdnHeader() error {
	data, err := json.Marshal(
		map[string]string{
			"device_id": inst.uuid,
		},
	)
	if err != nil {
		return err
	}
	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint:   urlMsisdnHeader,
			IsPost:     true,
			Connection: "keep-alive",
			Query:      generateSignature(b2s(data)),
		},
	)
	return err
}

func (inst *Instagram) contactPrefill() error {
	data, err := json.Marshal(
		map[string]string{
			"phone_id":   inst.pid,
			"_csrftoken": inst.token,
			"usage":      "prefill",
		},
	)
	if err != nil {
		return err
	}
	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint:   urlContactPrefill,
			IsPost:     true,
			Connection: "keep-alive",
			Query:      generateSignature(b2s(data)),
		},
	)
	return err
}

func (inst *Instagram) zrToken() error {
	_, err := inst.sendRequest(
		&reqOptions{
			Endpoint:   urlZrToken,
			IsPost:     false,
			Connection: "keep-alive",
			Query: map[string]string{
				"device_id":        inst.dID,
				"token_hash":       "",
				"custom_device_id": inst.uuid,
				"fetch_reason":     "token_expired",
			},
		},
	)
	return err
}

func (inst *Instagram) sendAdID() error {
	data, err := inst.prepareData(
		map[string]interface{}{
			"adid": inst.adid,
		},
	)
	if err != nil {
		return err
	}
	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint:   urlLogAttribution,
			IsPost:     true,
			Connection: "keep-alive",
			Query:      generateSignature(data),
		},
	)
	return err
}

func (inst *Instagram) Login() error {
	err := inst.readMsisdnHeader()
	if err != nil {
		return err
	}

	err = inst.syncFeatures()
	if err != nil {
		return err
	}

	err = inst.zrToken()
	if err != nil {
		return err
	}

	err = inst.sendAdID()
	if err != nil {
		return err
	}

	err = inst.contactPrefill()
	if err != nil {
		return err
	}

	result, err := json.Marshal(
		map[string]interface{}{
			"guid":                inst.uuid,
			"login_attempt_count": 0,
			"_csrftoken":          inst.token,
			"device_id":           inst.dID,
			"adid":                inst.adid,
			"phone_id":            inst.pid,
			"username":            inst.user,
			"password":            inst.pass,
			"google_tokens":       "[]",
		},
	)
	if err != nil {
		return err
	}
	body, err := inst.sendRequest(
		&reqOptions{
			Endpoint: urlLogin,
			Query:    generateSignature(b2s(result)),
			IsPost:   true,
			Login:    true,
		},
	)
	if err != nil {
		return err
	}
	inst.pass = ""

	res := accountResp{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	inst.Account = &res.Account
	inst.Account.inst = inst
	inst.rankToken = strconv.FormatInt(inst.Account.ID, 10) + "_" + inst.uuid
	inst.zrToken()

	return err
}

func (inst *Instagram) Logout() error {
	_, err := inst.sendSimpleRequest(urlLogout)
	inst.c.Jar = nil
	inst.c = nil
	return err
}

func (inst *Instagram) syncFeatures() error {
	data, err := inst.prepareData(
		map[string]interface{}{
			"id":          inst.uuid,
			"experiments": goInstaExperiments,
		},
	)
	if err != nil {
		return err
	}

	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint: urlQeSync,
			Query:    generateSignature(data),
			IsPost:   true,
			Login:    true,
		},
	)
	return err
}

func (inst *Instagram) megaphoneLog() error {
	data, err := inst.prepareData(
		map[string]interface{}{
			"id":        inst.Account.ID,
			"type":      "feed_aysf",
			"action":    "seen",
			"reason":    "",
			"device_id": inst.dID,
			"uuid":      generateMD5Hash(string(time.Now().Unix())),
		},
	)
	if err != nil {
		return err
	}
	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint: urlMegaphoneLog,
			Query:    generateSignature(data),
			IsPost:   true,
			Login:    true,
		},
	)
	return err
}

func (inst *Instagram) expose() error {
	data, err := inst.prepareData(
		map[string]interface{}{
			"id":         inst.Account.ID,
			"experiment": "ig_android_profile_contextual_feed",
		},
	)
	if err != nil {
		return err
	}

	_, err = inst.sendRequest(
		&reqOptions{
			Endpoint: urlExpose,
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)

	return err
}

func (inst *Instagram) GetMedia(o interface{}) (*FeedMedia, error) {
	media := &FeedMedia{
		inst:   inst,
		NextID: o,
	}
	return media, media.Sync()
}
