package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rbrick/linkmc/config"
	"log"
)

func WithVerifyChannel(channel string) Option {
	return func(bot Bot) {
		(bot).(*DiscordBot).VerifyChannel = channel
	}
}

type DiscordBot struct {
	CommandHandler
	conf config.Bot

	session *discordgo.Session

	channelNameToId map[string]string

	VerifyChannel string
}

func (b *DiscordBot) Init() error {
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

func (b *DiscordBot) Close() error {
	return b.session.Close()
}

func (b *DiscordBot) Config() config.Bot {
	return b.conf
}

func (b *DiscordBot) getChannelId(guildId, channelName string) string {
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
func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if b.getChannelId(m.GuildID, b.VerifyChannel) != m.ChannelID {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println("message")
	if m.Content[0] == '/' {
		b.CommandHandler.Handle(m.Content[1:])
	}
}

func NewDiscordBot(botConf config.Bot, options ...Option) Bot {
	b := &DiscordBot{
		CommandHandler:  NewCommandHandler(),
		conf:            botConf,
		channelNameToId: map[string]string{},
	}

	for _, opts := range options {
		opts(b)
	}

	return b
}
