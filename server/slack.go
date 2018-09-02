package main

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/model"
	"github.com/nlopes/slack"
)

func (p *SlackSubscriberPlugin) processSlackEvent(msg slack.RTMEvent, info *slack.Info, client *slack.Client) {
	p.API.LogDebug("Event Received")
	switch ev := msg.Data.(type) {
	case *slack.HelloEvent:
	case *slack.ConnectedEvent:
	case *slack.LatencyReport:
		// Ignore events
	case *slack.MessageEvent:
		if !isTargetChannel(ev.Msg.Channel, p.Slack.ChannelIDs) {
			return
		}
		if ev.Msg.User == "" {
			go p.postMessageWithAttachment(ev, info, client)
		} else {
			go p.postPlainMessage(ev, info)
		}
	case *slack.RTMError:
		fmt.Printf("Error: %s\n", ev.Error())
	case *slack.InvalidAuthEvent:
		fmt.Printf("Invalid credentials")
		return
	default:
		fmt.Printf("Unhandled event: %v", ev)
		// Ignore other events..
		// fmt.Printf("Unexpected: %v\n", msg.Data)
	}
}

func (p *SlackSubscriberPlugin) postPlainMessage(ev *slack.MessageEvent, info *slack.Info) {
	u := info.GetUserByID(ev.Msg.User)
	if _, err := p.API.CreatePost(&model.Post{
		ChannelId: p.Mattermost.ChannelID,
		UserId:    p.Mattermost.UserID,
		Message:   ev.Msg.Text,
		Type:      model.POST_DEFAULT,
		Props: model.StringInterface{
			"override_username": u.Name,
			"override_icon_url": u.Profile.Image48,
			"from_webhook":      "true",
		},
	}); err != nil {
		p.API.LogDebug("  Posg error: ", err.Error())
	} else {
		p.API.LogDebug("  Done")
	}
}

func (p *SlackSubscriberPlugin) postMessageWithAttachment(ev *slack.MessageEvent, info *slack.Info, client *slack.Client) {
	// timestamp := ev.Msg.Timestamp
	post, appErr := p.API.CreatePost(&model.Post{
		ChannelId: p.Mattermost.ChannelID,
		UserId:    p.Mattermost.UserID,
		Message:   ev.Msg.Text,
		Type:      model.POST_DEFAULT,
		Props: model.StringInterface{
			"override_username": ev.Msg.Username,
			"from_webhook":      "true",
		},
	})
	if appErr != nil {
		p.API.LogDebug("  Posg error: ", "details", appErr.Error())
		return
	}

	time.Sleep(3 * time.Second)
	h, err := client.GetChannelHistory(ev.Msg.Channel, slack.HistoryParameters{
		Latest: ev.Msg.Timestamp,
	})
	if err != nil {
		p.API.LogDebug("  Couldn't get channel history.", "Error", err.Error())
		return
	}
	if len(h.Messages) != 1 {
		p.API.LogDebug("  Couldn't get just one message from channel history.", "Attachment Count", len(h.Messages))
		return
	}
	msg := h.Messages[0]
	var attachments []*model.SlackAttachment
	for _, a := range msg.Attachments {
		attachments = append(attachments, &model.SlackAttachment{
			Fallback:   a.Fallback,
			AuthorName: a.AuthorName,
			AuthorLink: a.AuthorLink,
			AuthorIcon: a.AuthorIcon,
			Title:      "from Slack",
			TitleLink:  a.TitleLink,
			Text:       a.Text,
			ImageURL:   a.ImageURL,
			ThumbURL:   a.ThumbURL,
			Footer:     a.Footer,
			FooterIcon: a.FooterIcon,
		})
	}
	post.Props = model.StringInterface{
		"attachments":       attachments,
		"override_username": msg.Username,
		"override_icon_url": msg.Icons.IconEmoji,
		"from_webhook":      "true",
	}
	if _, appErr = p.API.UpdatePost(post); err != nil {
		p.API.LogDebug("  Post error: ", "details", appErr.Error())
		return
	}
}

func isTargetChannel(channelID string, targets []string) bool {
	for _, v := range targets {
		if v == channelID {
			return true
		}
	}
	return false
}
