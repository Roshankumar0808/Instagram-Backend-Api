package goinsta

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Comments struct {
	item     *Item
	endpoint string
	err      error

	Items                          []Comment       `json:"comments"`
	CommentCount                   int64           `json:"comment_count"`
	Caption                        Caption         `json:"caption"`
	CaptionIsEdited                bool            `json:"caption_is_edited"`
	HasMoreComments                bool            `json:"has_more_comments"`
	HasMoreHeadloadComments        bool            `json:"has_more_headload_comments"`
	ThreadingEnabled               bool            `json:"threading_enabled"`
	MediaHeaderDisplay             string          `json:"media_header_display"`
	InitiateAtTop                  bool            `json:"initiate_at_top"`
	InsertNewCommentToTop          bool            `json:"insert_new_comment_to_top"`
	PreviewComments                []Comment       `json:"preview_comments"`
	NextMaxID                      json.RawMessage `json:"next_max_id,omitempty"`
	NextMinID                      json.RawMessage `json:"next_min_id,omitempty"`
	CommentLikesEnabled            bool            `json:"comment_likes_enabled"`
	DisplayRealtimeTypingIndicator bool            `json:"display_realtime_typing_indicator"`
	Status                         string          `json:"status"`
}

func (comments *Comments) setValues() {
	for i := range comments.Items {
		comments.Items[i].setValues(comments.item.media.instagram())
	}
}

func newComments(item *Item) *Comments {
	c := &Comments{
		item: item,
	}
	return c
}

func (comments Comments) Error() error {
	return comments.err
}

func (comments *Comments) Disable() error {
	switch comments.item.media.(type) {
	case *StoryMedia:
		return fmt.Errorf("Incompatible type. Cannot use Disable() with StoryMedia type")
	default:
	}

	insta := comments.item.media.instagram()
	data, err := insta.prepareData(
		map[string]interface{}{
			"media_id": comments.item.ID,
		},
	)
	if err != nil {
		return err
	}

	_, err = insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlCommentDisable, comments.item.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	return err
}

func (comments *Comments) Enable() error {
	switch comments.item.media.(type) {
	case *StoryMedia:
		return fmt.Errorf("Incompatible type. Cannot use Enable() with StoryMedia type")
	default:
	}

	insta := comments.item.media.instagram()
	data, err := insta.prepareData(
		map[string]interface{}{
			"media_id": comments.item.ID,
		},
	)
	if err != nil {
		return err
	}

	_, err = insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlCommentEnable, comments.item.ID),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	return err
}

func (comments *Comments) Next() bool {
	if comments.err != nil {
		return false
	}

	item := comments.item
	insta := item.media.instagram()
	endpoint := comments.endpoint
	query := map[string]string{}
	if comments.NextMaxID != nil {
		next, _ := strconv.Unquote(string(comments.NextMaxID))
		query["max_id"] = next
	} else if comments.NextMinID != nil {
		next, _ := strconv.Unquote(string(comments.NextMinID))
		query["min_id"] = next
	}

	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint:   endpoint,
			Connection: "keep-alive",
			Query:      query,
		},
	)
	if err == nil {
		c := Comments{}
		err = json.Unmarshal(body, &c)
		if err == nil {
			*comments = c
			comments.endpoint = endpoint
			comments.item = item
			if (!comments.HasMoreComments || comments.NextMaxID == nil) &&
				(!comments.HasMoreHeadloadComments || comments.NextMinID == nil) {
				comments.err = ErrNoMore
			}
			comments.setValues()
			return true
		}
	}
	comments.err = err
	return false
}

func (comments *Comments) Sync() {
	endpoint := fmt.Sprintf(urlCommentSync, comments.item.ID)
	comments.endpoint = endpoint
	return
}

func (comments *Comments) Add(text string) (err error) {
	var opt *reqOptions
	item := comments.item
	insta := item.media.instagram()

	switch item.media.(type) {
	case *StoryMedia:
		to, err := prepareRecipients(item)
		if err != nil {
			return err
		}

		query := insta.prepareDataQuery(
			map[string]interface{}{
				"recipient_users": to,
				"action":          "send_item",
				"media_id":        item.ID,
				"client_context":  generateUUID(),
				"text":            text,
				"entry":           "reel",
				"reel_id":         item.User.ID,
			},
		)
		opt = &reqOptions{
			Connection: "keep-alive",
			Endpoint:   fmt.Sprintf("%s?media_type=%s", urlReplyStory, item.MediaToString()),
			Query:      query,
			IsPost:     true,
		}
	case *FeedMedia:
		var data string
		data, err = insta.prepareData(
			map[string]interface{}{
				"comment_text": text,
			},
		)
		opt = &reqOptions{
			Endpoint: fmt.Sprintf(urlCommentAdd, item.Pk),
			Query:    generateSignature(data),
			IsPost:   true,
		}
	}
	if err != nil {
		return err
	}

	_, err = insta.sendRequest(opt)
	return err
}

func (comments *Comments) Del(comment *Comment) error {
	insta := comments.item.media.instagram()

	data, err := insta.prepareData()
	if err != nil {
		return err
	}
	id := comment.getid()

	_, err = insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlCommentDelete, comments.item.ID, id),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	return err
}

func (comments *Comments) DelByID(id string) error {
	return comments.Del(&Comment{idstr: id})
}

func (comments *Comments) DelMine(limit int) error {
	i := 0
	if limit <= 0 {
		i = limit - 1
	}
	comments.Sync()

	insta := comments.item.media.instagram()
floop:
	for comments.Next() {
		for _, c := range comments.Items {
			if c.UserID == insta.Account.ID || c.User.ID == insta.Account.ID {
				if i >= limit {
					break floop
				}
				comments.Del(&c)
				i++
			}
		}
	}
	if err := comments.Error(); err != nil && err != ErrNoMore {
		return err
	}
	return nil
}

type Comment struct {
	inst  *Instagram
	idstr string

	ID                             int64     `json:"pk"`
	Text                           string    `json:"text"`
	Type                           int       `json:"type"`
	User                           User      `json:"user"`
	UserID                         int64     `json:"user_id"`
	BitFlags                       int       `json:"bit_flags"`
	ChildCommentCount              int       `json:"child_comment_count"`
	CommentIndex                   int       `json:"comment_index"`
	CommentLikeCount               int       `json:"comment_like_count"`
	ContentType                    string    `json:"content_type"`
	CreatedAt                      int64     `json:"created_at"`
	CreatedAtUtc                   int64     `json:"created_at_utc"`
	DidReportAsSpam                bool      `json:"did_report_as_spam"`
	HasLikedComment                bool      `json:"has_liked_comment"`
	InlineComposerDisplayCondition string    `json:"inline_composer_display_condition"`
	OtherPreviewUsers              []User    `json:"other_preview_users"`
	PreviewChildComments           []Comment `json:"preview_child_comments"`
	NextMaxChildCursor             string    `json:"next_max_child_cursor,omitempty"`
	HasMoreTailChildComments       bool      `json:"has_more_tail_child_comments,omitempty"`
	NextMinChildCursor             string    `json:"next_min_child_cursor,omitempty"`
	HasMoreHeadChildComments       bool      `json:"has_more_head_child_comments,omitempty"`
	NumTailChildComments           int       `json:"num_tail_child_comments,omitempty"`
	NumHeadChildComments           int       `json:"num_head_child_comments,omitempty"`
	Status                         string    `json:"status"`
}

func (c *Comment) setValues(inst *Instagram) {
	c.User.inst = inst
	for i := range c.OtherPreviewUsers {
		c.OtherPreviewUsers[i].inst = inst
	}
	for i := range c.PreviewChildComments {
		c.PreviewChildComments[i].setValues(inst)
	}
}

func (c Comment) getid() string {
	switch {
	case c.ID == 0:
		return c.idstr
	case c.idstr == "":
		return strconv.FormatInt(c.ID, 10)
	}
	return ""
}

func (c *Comment) Like() error {
	data, err := c.inst.prepareData()
	if err != nil {
		return err
	}

	_, err = c.inst.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlCommentLike, c.getid()),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	return err
}

func (c *Comment) Unlike() error {
	data, err := c.inst.prepareData()
	if err != nil {
		return err
	}

	_, err = c.inst.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlCommentUnlike, c.getid()),
			Query:    generateSignature(data),
			IsPost:   true,
		},
	)
	return err
}
