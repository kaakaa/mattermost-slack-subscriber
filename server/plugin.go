package main

import (
	"context"

	"github.com/mattermost/mattermost-server/plugin"
	"github.com/nlopes/slack"
)

type SlackSubscriberPlugin struct {
	plugin.MattermostPlugin
	Mattermost *MattermostSettings
	Slack      *SlackSettings
	ctx        context.Context
	cancel     context.CancelFunc
}
type MattermostSettings struct {
	ChannelID string
	UserID    string
}

type SlackSettings struct {
	AccessToken string
	BotToken    string
	ChannelIDs  []string
}

func (p *SlackSubscriberPlugin) OnDeactivate() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *SlackSubscriberPlugin) connect() {
	if p.cancel != nil {
		p.cancel()
	}
	ctx, cancel := context.WithCancel(p.ctx)
	p.cancel = cancel

	api := slack.New(p.Slack.BotToken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	p.API.LogInfo("Strat connecting")
	for {
		select {
		case <-ctx.Done():
			p.API.LogInfo("Disconnect RTM")
			if err := rtm.Disconnect(); err != nil {
				p.API.LogError(err.Error())
			}
			return
		case msg := <-rtm.IncomingEvents:
			p.processSlackEvent(msg, rtm.GetInfo(), slack.New(p.Slack.AccessToken))
		}
	}
}
