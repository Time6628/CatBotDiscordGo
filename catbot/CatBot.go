package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"flag"
	"time"
	"regexp"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"errors"
	"bytes"
	"github.com/Time6628/OpenTDB-Go"
	"github.com/nanobox-io/golang-scribble"
	"os"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var (
	token string
	BotID string
	client = fasthttp.Client{ReadTimeout: time.Second * 10, WriteTimeout: time.Second * 10}
	trivia = OpenTDB_Go.New(client)
	triviaRunning = false
	db *scribble.Driver
)

func main()  {
	go forever()
	fmt.Println("Starting Catbot 2.0")

	if token == "" {
		fmt.Println("No token provided. Please run: catbot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)

	u, err := dg.User("@me")
	if err != nil {
		panic(err)
	}
	BotID = u.ID

	err = os.Mkdir("./database", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	db, err = scribble.New("./database", &scribble.Options{})
	if err != nil {
		panic(err)
	}

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildJoin)

	fmt.Println("CatBot is now running.  Press CTRL-C to exit.")
	select {}
}

func forever() {}

func guildJoin(s *discordgo.Session, g *discordgo.GuildMemberAdd) {
	if isMuted(g.User.ID, g.GuildID) {
		channels, _ := s.GuildChannels(g.GuildID)
		for _, channel := range channels {
			if !alreadyMutedInChannel(g.User.ID, channel) {
				s.ChannelPermissionSet(channel.ID, g.User.ID, "member", 0, discordgo.PermissionSendMessages)
			}
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	d, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}

	g, err := s.Guild(d.GuildID)
	if err != nil {
		return
	}

	member, err := s.GuildMember(g.ID, m.Author.ID)
	if err != nil {
		return
	}

	roles := member.Roles

	c := strings.ToLower(m.Content)

	filters := []*regexp.Regexp{regexp.MustCompile("dick"), regexp.MustCompile("fuck"), regexp.MustCompile("penis"), regexp.MustCompile("vagina"), regexp.MustCompile("fag"), regexp.MustCompile("\\brape"), regexp.MustCompile("slut"), regexp.MustCompile("slut"), regexp.MustCompile("hitler"), regexp.MustCompile("\\b(jack)?ass(holes?|lick|wipe)?\\b"), regexp.MustCompile("arse(hole)?"), regexp.MustCompile("bitch"), regexp.MustCompile("whore"), regexp.MustCompile("nigg(er|a)"), regexp.MustCompile("bastard"), regexp.MustCompile("bea?stiality"), regexp.MustCompile("negro"), regexp.MustCompile("retard"), regexp.MustCompile("\\bcum\\b"), regexp.MustCompile("cunt"), regexp.MustCompile("dildo"), regexp.MustCompile("bollocks?"), regexp.MustCompile("\\bwank"), regexp.MustCompile("jizz"), regexp.MustCompile("piss"),}
	filter := false


	admin := false
	for i := 0; i < len(roles); i++ {
		role, _ := s.State.Role(g.ID, roles[i])
		if (role.Permissions & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator {
			admin = true
		}
	}

	if IsChannelFiltered(d.ID, d.GuildID) {
		for _, filt := range filters {
			if filt.MatchString(c) {
				filter = true
			}
		}
	}

	if filter {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Messaged removed from <@" + m.Author.ID + ">.")
		removeLater(s, rm)
		return
	}

	if !strings.HasPrefix(c, "!") {
		return
	}

	cmdBits := strings.Split(c, " ")

	switch cmdBits[0] {
	case helpCmd.Prefix:
		helpCmd.Function.(func(*discordgo.Session, *discordgo.User, bool))(s, m.Author, admin)
	case removeFilterCmd.Prefix:
		if !admin {
			return
		}
		removeFilterCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message))(s, d, m.Message)
	case enableFilterCmd.Prefix:
		if !admin {
			return
		}
		enableFilterCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message))(s, d, m.Message)
	case infoCmd.Prefix:
		infoCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case catbotCmd.Prefix:
		catbotCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case muteCmd.Prefix:
		if !admin {
			return
		}
		if !strings.Contains(c, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		user_id := strings.TrimPrefix(strings.TrimSuffix(cmdBits[1], ">"), "<@")
		muteCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, string))(s, d, m.Message, user_id)
	case allMuteCmd.Prefix:
		if !admin {
			return
		}
		if !strings.Contains(c, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		user_id := strings.TrimPrefix(strings.TrimSuffix(cmdBits[1], ">"), "<@")
		allMuteCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, string))(s, d, m.Message, user_id)
	case donationHelpCmd.Prefix:
		donationHelpCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case catCmd.Prefix:
		catCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case snekCmd.Prefix:
		snekCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case broomCmd.Prefix:
		broomCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case rickCmd.Prefix:
		rickCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case vktrsCmd.Prefix:
		vktrsCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case clearCmd.Prefix:
		if canManageMessage(s, m.Author, d) {
			return
		}
		clearCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, *discordgo.Member, []string))(s, d, m.Message, member, cmdBits)
	case triviaCmd.Prefix:
		triviaCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case topicCmd.Prefix:
		topicCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	}
}

func doLater(i func()) {
	timer := time.NewTimer(time.Minute * 1)
	<- timer.C
	i()
}

func countChannels(guilds []*discordgo.Guild) (channels int) {
	for _, guild := range guilds {
		channels = len(guild.Channels) + channels
	}
	return
}

func countUsers(guilds []*discordgo.Guild) (users int) {
	for _, guild := range guilds {
		users = guild.MemberCount + users
	}
	return
}

func formatError(err error) string {
	return "```" + err.Error() + "```"
}

func canManageMessage(session *discordgo.Session, user *discordgo.User, channel *discordgo.Channel) bool {
	uPerms, _ := session.UserChannelPermissions(user.ID, channel.ID)
	if (uPerms&discordgo.PermissionManageMessages) == discordgo.PermissionManageMessages {
		return true
	}
	return false
}

func clearChannelChat(i int, channel *discordgo.Channel, session *discordgo.Session) {
	fmt.Println("Clearing channel messages...")
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not get messages.")
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	todelete := []string{}
	for _, message := range messages {
		todelete = append(todelete, message.ID)
	}
	session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, err := session.ChannelMessageSend(channel.ID, "Messages removed in channel " + channel.Name)
	if err != nil {
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	removeLater(session, m)
}

func clearUserChat(i int, channel *discordgo.Channel, session *discordgo.Session, id string) {
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not get messages.")
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	todelete := []string{}
	for _, message := range messages {
		if message.Author.ID == id {
			todelete = append(todelete, message.ID)
		}
	}
	session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, _ := session.ChannelMessageSend(channel.ID, "Messages removed for user <@" + id + "> in channel " + channel.Name)
	removeLater(session, m)
}

func removeLaterBulk(session *discordgo.Session, messages []*discordgo.Message) {
	for _, z := range messages {
		timer := time.NewTimer(time.Second * 5)
		<- timer.C
		session.ChannelMessageDelete(z.ChannelID, z.ID)
	}
}

func alreadyMutedInChannel(id string, channel *discordgo.Channel) (b bool) {
	permissions := channel.PermissionOverwrites
	for i := 0; i < len(permissions); i++ {
		permission := permissions[i]
		if permission.ID == id && permission.Type == "member" {
			b = permission.Deny == discordgo.PermissionSendMessages
		}
	}
	return
}

func removeLater(s *discordgo.Session, m *discordgo.Message) {
	timer := time.NewTimer(time.Second * 5)
	<- timer.C
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

type CatResponse struct {
	URL string `json:"file"`
}

func getJson(url string, target interface{}) error {
	stat, body, err := client.Get(nil, url)
	if err != nil || stat != 200 {
		return errors.New("Could not obtain json response")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(target)
}

type UnfilteredChannel struct {
	ChannelID string `json:"ID"`
}

type MutedUser struct {
	DiscordID string `json:"ID"`
}


func addToUnfilterd(channel_id, guild_id string) {
	channel := UnfilteredChannel{ChannelID:channel_id}
	if err := db.Write(guild_id, channel_id, channel); err != nil {
		panic(err)
	}
}

func removeFromUnfiltered(channel_id, guild_id string) (err error) {
	if err = db.Delete(guild_id, channel_id); err != nil {
		return
	}
	return
}

func IsChannelFiltered(channel_id, guild_id string) (b bool) {
	c := UnfilteredChannel{}
	b = true
	if err := db.Read(guild_id, channel_id, &c); err != nil {
		return
	}
	b = false
	return
}

func addToMuted(user_id, guild_id string) {
	user := MutedUser{DiscordID:user_id}
	if err := db.Write(guild_id, user_id, user); err != nil {
		panic(err)
	}
}

func removeFromMuted(user_id, guild_id string) (err error) {
	if err = db.Delete(guild_id, user_id); err != nil {
		return
	}
	return
}

func isMuted(user_id, guild_id string) bool {
	user := MutedUser{}

	if err := db.Read(guild_id, user_id, &user); err != nil {
		return false
	}

	return true
}