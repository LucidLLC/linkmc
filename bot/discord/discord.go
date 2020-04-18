package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rbrick/linkmc/bot"
	"github.com/rbrick/linkmc/config"
	"log"
)

type Bot struct {
	bot.CommandHandler
	conf config.Bot

	session *discordgo.Session

	channelNameToId map[string]string

	verifyChannel string
}

func (b *Bot) Init() error {
	log.Println("creating discord bot")
	discord, err := discordgo.New("Bot " + b.conf.Token)

	if err != nil {
		return err
	}

	log.Println("registering discord handlers")
	discord.AddHandler(b.messageCreate)

	log.Println("opening discord websockets")
	err = discord.Open()

	if err != nil {
		return err
	}

	log.Println("discord bot now running")
	b.session = discord
	return nil
}

func (b *Bot) Close() error {
	return b.session.Close()
}

func (b *Bot) Config() config.Bot {
	return b.conf
}

func (b *Bot) getChannelId(guildId, channelName string) string {
	key := guildId + ":" + channelName
	if v, ok := b.channelNameToId[key]; ok {
		return v
	}

	ch, err := b.session.GuildChannels(guildId)

	if err != nil {
		return ""
	}

	for _, channel := range ch {
		if channel.Name == channelName {
			b.channelNameToId[key] = channel.ID
		}
	}

	return b.channelNameToId[key]
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if b.getChannelId(m.GuildID, b.verifyChannel) != m.ChannelID {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println("message")
	if m.Content[0] == '/' {
		b.CommandHandler.Receive(m.Content[1:])
	}
}

func NewDiscordBot(verifyChannel string, botConf config.Bot) bot.Bot {
	return &Bot{
		CommandHandler:  bot.NewCommandHandler(),
		conf:            botConf,
		verifyChannel:   verifyChannel,
		channelNameToId: map[string]string{},
	}
}
