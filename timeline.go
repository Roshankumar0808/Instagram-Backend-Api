package goinsta

import (
	"encoding/json"
)


type Timeline struct {
	inst *Instagram
}

func newTimeline(inst *Instagram) *Timeline {
	time := &Timeline{
		inst: inst,
	}
	return time
}


func (time *Timeline) Get() *FeedMedia {
	insta := time.inst
	media := &FeedMedia{}
	media.inst = insta
	media.endpoint = urlTimeline
	return media
}


func (time *Timeline) Stories() (*Tray, error) {
	body, err := time.inst.sendSimpleRequest(urlStories)
	if err == nil {
		tray := &Tray{}
		err = json.Unmarshal(body, tray)
		if err != nil {
			return nil, err
		}
		tray.set(time.inst, urlStories)
		return tray, nil
	}
	return nil, err
}
