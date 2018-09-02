package main

import (
	"context"
	"strings"
)

type Config struct {
	SlackOAuthAccessToken string
	SlackBotUserToken     string
	SlackChannels         string
	MattermostChannelID   string
	MattermostUserID      string
}

func (p *SlackSubscriberPlugin) OnConfigurationChange() error {
	c := &Config{}
	if err := p.API.LoadPluginConfiguration(c); err != nil {
		return err
	}
	p.Mattermost = &MattermostSettings{
		ChannelID: c.MattermostChannelID,
		UserID:    c.MattermostUserID,
	}
	p.Slack = &SlackSettings{
		AccessToken: c.SlackOAuthAccessToken,
		BotToken:    c.SlackBotUserToken,
		ChannelIDs:  strings.Split(c.SlackChannels, ","),
	}
	p.ctx = context.Background()
	go p.connect()
	return nil
}
