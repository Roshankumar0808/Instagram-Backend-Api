package goinsta

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Users struct {
	inst *Instagram

	err      error
	endpoint string

	Status    string          `json:"status"`
	BigList   bool            `json:"big_list"`
	Users     []User          `json:"users"`
	PageSize  int             `json:"page_size"`
	RawNextID json.RawMessage `json:"next_max_id"`
	NextID    string          `json:"-"`
}

func newUsers(inst *Instagram) *Users {
	users := &Users{inst: inst}

	return users
}

func (users *Users) SetInstagram(inst *Instagram) {
	users.inst = inst
}

var ErrNoMore = errors.New("List end have been reached")

func (users *Users) Next() bool {
	if users.err != nil {
		return false
	}

	insta := users.inst
	endpoint := users.endpoint

	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: endpoint,
			Query: map[string]string{
				"max_id":             users.NextID,
				"ig_sig_key_version": goInstaSigKeyVersion,
				"rank_token":         insta.rankToken,
			},
		},
	)
	if err == nil {
		usrs := Users{}
		err = json.Unmarshal(body, &usrs)
		if err == nil {
			if len(usrs.RawNextID) > 0 && usrs.RawNextID[0] == '"' && usrs.RawNextID[len(usrs.RawNextID)-1] == '"' {
				if err := json.Unmarshal(usrs.RawNextID, &usrs.NextID); err != nil {
					users.err = err
					return false
				}
			} else {
				var nextID int64
				if err := json.Unmarshal(usrs.RawNextID, &nextID); err != nil {
					users.err = err
					return false
				}
				usrs.NextID = strconv.FormatInt(nextID, 10)
			}
			*users = usrs
			if !usrs.BigList || usrs.NextID == "" {
				users.err = ErrNoMore
			}
			users.inst = insta
			users.endpoint = endpoint
			users.setValues()
			return true
		}
	}
	users.err = err
	return false
}

func (users *Users) Error() error {
	return users.err
}

func (users *Users) setValues() {
	for i := range users.Users {
		users.Users[i].inst = users.inst
	}
}

type userResp struct {
	Status string `json:"status"`
	User   User   `json:"user"`
}

type User struct {
	inst *Instagram

	ID                         int64   `json:"pk"`
	Username                   string  `json:"username"`
	FullName                   string  `json:"full_name"`
	Biography                  string  `json:"biography"`
	ProfilePicURL              string  `json:"profile_pic_url"`
	Email                      string  `json:"email"`
	PhoneNumber                string  `json:"phone_number"`
	IsBusiness                 bool    `json:"is_business"`
	Gender                     int     `json:"gender"`
	ProfilePicID               string  `json:"profile_pic_id"`
	HasAnonymousProfilePicture bool    `json:"has_anonymous_profile_picture"`
	IsPrivate                  bool    `json:"is_private"`
	IsUnpublished              bool    `json:"is_unpublished"`
	AllowedCommenterType       string  `json:"allowed_commenter_type"`
	IsVerified                 bool    `json:"is_verified"`
	MediaCount                 int     `json:"media_count"`
	FollowerCount              int     `json:"follower_count"`
	FollowingCount             int     `json:"following_count"`
	FollowingTagCount          int     `json:"following_tag_count"`
	MutualFollowersID          []int64 `json:"profile_context_mutual_follow_ids"`
	ProfileContext             string  `json:"profile_context"`
	GeoMediaCount              int     `json:"geo_media_count"`
	ExternalURL                string  `json:"external_url"`
	HasBiographyTranslation    bool    `json:"has_biography_translation"`
	ExternalLynxURL            string  `json:"external_lynx_url"`
	BiographyWithEntities      struct {
		RawText  string        `json:"raw_text"`
		Entities []interface{} `json:"entities"`
	} `json:"biography_with_entities"`
	UsertagsCount                int          `json:"usertags_count"`
	HasChaining                  bool         `json:"has_chaining"`
	IsFavorite                   bool         `json:"is_favorite"`
	IsFavoriteForStories         bool         `json:"is_favorite_for_stories"`
	IsFavoriteForHighlights      bool         `json:"is_favorite_for_highlights"`
	CanBeReportedAsFraud         bool         `json:"can_be_reported_as_fraud"`
	ShowShoppableFeed            bool         `json:"show_shoppable_feed"`
	ShoppablePostsCount          int          `json:"shoppable_posts_count"`
	ReelAutoArchive              string       `json:"reel_auto_archive"`
	HasHighlightReels            bool         `json:"has_highlight_reels"`
	PublicEmail                  string       `json:"public_email"`
	PublicPhoneNumber            string       `json:"public_phone_number"`
	PublicPhoneCountryCode       string       `json:"public_phone_country_code"`
	ContactPhoneNumber           string       `json:"contact_phone_number"`
	CityID                       int64        `json:"city_id"`
	CityName                     string       `json:"city_name"`
	AddressStreet                string       `json:"address_street"`
	DirectMessaging              string       `json:"direct_messaging"`
	Latitude                     float64      `json:"latitude"`
	Longitude                    float64      `json:"longitude"`
	Category                     string       `json:"category"`
	BusinessContactMethod        string       `json:"business_contact_method"`
	IncludeDirectBlacklistStatus bool         `json:"include_direct_blacklist_status"`
	HdProfilePicURLInfo          PicURLInfo   `json:"hd_profile_pic_url_info"`
	HdProfilePicVersions         []PicURLInfo `json:"hd_profile_pic_versions"`
	School                       School       `json:"school"`
	Byline                       string       `json:"byline"`
	SocialContext                string       `json:"social_context,omitempty"`
	SearchSocialContext          string       `json:"search_social_context,omitempty"`
	MutualFollowersCount         float64      `json:"mutual_followers_count"`
	LatestReelMedia              int64        `json:"latest_reel_media,omitempty"`
	IsCallToActionEnabled        bool         `json:"is_call_to_action_enabled"`
	FbPageCallToActionID         string       `json:"fb_page_call_to_action_id"`
	Zip                          string       `json:"zip"`
	Friendship                   Friendship   `json:"friendship_status"`
}

func (user *User) SetInstagram(insta *Instagram) {
	user.inst = insta
}

func (inst *Instagram) NewUser() *User {
	return &User{inst: inst}
}

func (user *User) Sync(params ...interface{}) error {
	insta := user.inst
	body, err := insta.sendSimpleRequest(urlUserInfo, user.ID)
	if err == nil {
		resp := userResp{}
		err = json.Unmarshal(body, &resp)
		if err == nil {
			*user = resp.User
			user.inst = insta
			for _, param := range params {
				switch b := param.(type) {
				case bool:
					if b {
						err = user.FriendShip()
					}
				}
			}
		}
	}
	return err
}

func (user *User) Following() *Users {
	users := &Users{}
	users.inst = user.inst
	users.endpoint = fmt.Sprintf(urlFollowing, user.ID)
	return users
}

func (user *User) Followers() *Users {
	users := &Users{}
	users.inst = user.inst
	users.endpoint = fmt.Sprintf(urlFollowers, user.ID)
	return users
}

func (user *User) Block() error {
	insta := user.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"user_id": user.ID,
		},
	)
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserBlock, user.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	if err != nil {
		return err
	}
	resp := friendResp{}
	err = json.Unmarshal(body, &resp)
	user.Friendship = resp.Friendship
	if err != nil {
		return err
	}

	return nil
}

func (user *User) Unblock() error {
	insta := user.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"user_id": user.ID,
		},
	)
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserUnblock, user.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	if err != nil {
		return err
	}
	resp := friendResp{}
	err = json.Unmarshal(body, &resp)
	user.Friendship = resp.Friendship
	if err != nil {
		return err
	}

	return nil
}

func (user *User) Mute(opt muteOption) error {
	return muteOrUnmute(user, opt, urlUserMute)
}

func (user *User) Unmute(opt muteOption) error {
	return muteOrUnmute(user, opt, urlUserUnmute)
}

func muteOrUnmute(user *User, opt muteOption, endpoint string) error {
	insta := user.inst
	data, err := insta.prepareData(
		generateMuteData(user, opt),
	)
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: endpoint,
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	if err != nil {
		return err
	}
	resp := friendResp{}
	err = json.Unmarshal(body, &resp)
	user.Friendship = resp.Friendship
	if err != nil {
		return err
	}

	return nil
}

func generateMuteData(user *User, opt muteOption) map[string]interface{} {
	data := map[string]interface{}{
		"user_id": user.ID,
	}

	switch opt {
	case MuteAll:
		data["target_reel_author_id"] = user.ID
		data["target_posts_author_id"] = user.ID
	case MuteStory:
		data["target_reel_author_id"] = user.ID
	case MuteFeed:
		data["target_posts_author_id"] = user.ID
	}

	return data
}

func (user *User) Follow() error {
	insta := user.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"user_id": user.ID,
		},
	)
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserFollow, user.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	if err != nil {
		return err
	}
	resp := friendResp{}
	err = json.Unmarshal(body, &resp)
	user.Friendship = resp.Friendship
	if err != nil {
		return err
	}

	return nil
}

func (user *User) Unfollow() error {
	insta := user.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"user_id": user.ID,
		},
	)
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserUnfollow, user.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	if err != nil {
		return err
	}
	resp := friendResp{}
	err = json.Unmarshal(body, &resp)
	user.Friendship = resp.Friendship
	if err != nil {
		return err
	}

	return nil
}

func (user *User) FriendShip() error {
	insta := user.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"user_id": user.ID,
		},
	)
	if err != nil {
		return err
	}

	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlFriendship, user.ID),
			Query:    generateSignature(data),
		},
	)
	if err == nil {
		err = json.Unmarshal(body, &user.Friendship)
	}
	return err
}

func (user *User) Feed(params ...interface{}) *FeedMedia {
	insta := user.inst

	media := &FeedMedia{}
	media.inst = insta
	media.endpoint = urlUserFeed
	media.uid = user.ID

	for _, param := range params {
		switch s := param.(type) {
		case string:
			media.timestamp = s
		}
	}

	return media
}

func (user *User) Stories() *StoryMedia {
	media := &StoryMedia{}
	media.uid = user.ID
	media.inst = user.inst
	media.endpoint = urlUserStories
	return media
}

func (user *User) Highlights() ([]StoryMedia, error) {
	query := []trayRequest{
		{"SUPPORTED_SDK_VERSIONS", "9.0,10.0,11.0,12.0,13.0,14.0,15.0,16.0,17.0,18.0,19.0,20.0,21.0,22.0,23.0,24.0"},
		{"FACE_TRACKER_VERSION", "10"},
		{"segmentation", "segmentation_enabled"},
		{"COMPRESSION", "ETC2_COMPRESSION"},
	}
	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	body, err := user.inst.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserHighlights, user.ID),
			Query:    generateSignature(b2s(data)),
		},
	)
	if err == nil {
		tray := &Tray{}
		err = json.Unmarshal(body, &tray)
		if err == nil {
			tray.set(user.inst, "")
			for i := range tray.Stories {
				if len(tray.Stories[i].Items) == 0 {
					err = tray.Stories[i].Sync()
					if err != nil {
						return nil, err
					}
				}
			}
			return tray.Stories, nil
		}
	}
	return nil, err
}

func (user *User) Tags(minTimestamp []byte) (*FeedMedia, error) {
	timestamp := b2s(minTimestamp)
	body, err := user.inst.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlUserTags, user.ID),
			Query: map[string]string{
				"max_id":         "",
				"rank_token":     user.inst.rankToken,
				"min_timestamp":  timestamp,
				"ranked_content": "true",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	media := &FeedMedia{}
	err = json.Unmarshal(body, media)
	media.inst = user.inst
	media.endpoint = urlUserTags
	media.uid = user.ID
	return media, err
}
