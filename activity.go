package goinsta

import (
	"encoding/json"
	"strconv"
)

type Activity struct {
	inst *Instagram
}

type FollowingActivity struct {
	inst *Instagram
	err  error

	AutoLoadMoreEnabled bool  `json:"auto_load_more_enabled"`
	NextID              int64 `json:"next_max_id"`
	Stories             []struct {
		Type      int `json:"type"`
		StoryType int `json:"story_type"`
		Args      struct {
			MediaDestination string `json:"media_destination"`
			Destination      string `json:"destination"`
			Text             string `json:"text"`
			Links            []struct {
				Start int    `json:"start"`
				End   int    `json:"end"`
				Type  string `json:"type"`
				ID    string `json:"id"`
			} `json:"links"`
			ProfileID               int64  `json:"profile_id"`
			ProfileImage            string `json:"profile_image"`
			SecondProfileID         int64  `json:"second_profile_id"`
			SecondProfileImage      string `json:"second_profile_image"`
			ProfileImageDestination string `json:"profile_image_destination"`
			Media                   []struct {
				ID    string `json:"id"`
				Image string `json:"image"`
			} `json:"media"`
			Timestamp int64  `json:"timestamp"`
			Tuuid     string `json:"tuuid"`
		} `json:"args"`
		Counts struct {
		} `json:"counts"`
		Pk string `json:"pk"`
	} `json:"stories"`
	Status string `json:"status"`
}

func (act *FollowingActivity) Error() error {
	return act.err
}

func (act *FollowingActivity) Next() bool {
	if act.err != nil {
		return false
	}
	insta := act.inst
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: urlActivityFollowing,
			Query: map[string]string{
				"max_id": strconv.FormatInt(act.NextID, 10),
			},
			IsPost: false,
		},
	)
	if err == nil {
		act2 := FollowingActivity{}
		err = json.Unmarshal(body, &act2)
		if err == nil {
			*act = act2
			act.inst = insta
			if len(act.Stories) == 0 || act.NextID == 0 {
				act.err = ErrNoMore
			}
			return true
		}
	}
	act.err = err
	return false
}

type MineActivity struct {
	inst *Instagram
	err  error

	Ad struct {
		Items []struct {
			Algorithm       string        `json:"algorithm"`
			SocialContext   string        `json:"social_context"`
			Icon            string        `json:"icon"`
			Caption         string        `json:"caption"`
			MediaIds        []interface{} `json:"media_ids"`
			ThumbnailUrls   []interface{} `json:"thumbnail_urls"`
			LargeUrls       []interface{} `json:"large_urls"`
			MediaInfos      []interface{} `json:"media_infos"`
			Value           float64       `json:"value"`
			IsNewSuggestion bool          `json:"is_new_suggestion"`
		} `json:"items"`
		MoreAvailable bool `json:"more_available"`
	} `json:"aymf"`
	Counts struct {
		PhotosOfYou int `json:"photos_of_you"`
		Requests    int `json:"requests"`
	} `json:"counts"`
	FriendRequestStories []interface{} `json:"friend_request_stories"`
	Stories              []struct {
		Type      int `json:"type"`
		StoryType int `json:"story_type"`
		Args      struct {
			Text  string `json:"text"`
			Links []struct {
				Start int    `json:"start"`
				End   int    `json:"end"`
				Type  string `json:"type"`
				ID    string `json:"id"`
			} `json:"links"`
			InlineFollow struct {
				UserInfo        User `json:"user_info"`
				Following       bool `json:"following"`
				OutgoingRequest bool `json:"outgoing_request"`
			} `json:"inline_follow"`
			Actions         []string `json:"actions"`
			ProfileID       int64    `json:"profile_id"`
			ProfileImage    string   `json:"profile_image"`
			Timestamp       float64  `json:"timestamp"`
			Tuuid           string   `json:"tuuid"`
			Clicked         bool     `json:"clicked"`
			ProfileName     string   `json:"profile_name"`
			LatestReelMedia int64    `json:"latest_reel_media"`
		} `json:"args"`
		Counts struct {
		} `json:"counts"`
		Pk string `json:"pk"`
	} `json:"old_stories"`
	ContinuationToken int64       `json:"continuation_token"`
	Subscription      interface{} `json:"subscription"`
	NextID            int64       `json:"next_max_id"`
	Status            string      `json:"status"`
}

func (act *MineActivity) Error() error {
	return act.err
}

func (act *MineActivity) Next() bool {
	if act.err != nil {
		return false
	}
	insta := act.inst
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: urlActivityRecent,
			Query: map[string]string{
				"max_id": strconv.FormatInt(act.NextID, 10),
			},
			IsPost: false,
		},
	)
	if err == nil {
		act2 := MineActivity{}
		err = json.Unmarshal(body, &act2)
		if err == nil {
			*act = act2
			act.inst = insta
			if len(act.Stories) == 0 || act.NextID == 0 {
				act.err = ErrNoMore
			}
			return true
		}
	}
	act.err = err
	return false
}

func newActivity(inst *Instagram) *Activity {
	act := &Activity{
		inst: inst,
	}
	return act
}

func (act *Activity) Following() *FollowingActivity {
	insta := act.inst
	nact := &FollowingActivity{inst: insta}
	return nact
}

func (act *Activity) Recent() *MineActivity {
	insta := act.inst
	nact := &MineActivity{inst: insta}
	return nact
}
