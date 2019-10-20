package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Time6628/OpenTDB-Go"
	"github.com/bwmarrin/discordgo"
	"github.com/nanobox-io/golang-scribble"
	"github.com/valyala/fasthttp"
	"os"
	"regexp"
	"strings"
	"time"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var (
	token         string
	BotID         string
	client        = fasthttp.Client{ReadTimeout: time.Second * 10, WriteTimeout: time.Second * 10}
	trivia        = OpenTDB_Go.New(client)
	triviaRunning = false
	db            *scribble.Driver
)

func main() {
	go forever()
	fmt.Println("Starting Catbot 2.1")

	if token == "" {
		fmt.Println("No token provided. Please run: catbot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

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
				_ = s.ChannelPermissionSet(channel.ID, g.User.ID, "member", 0, discordgo.PermissionSendMessages)
			}
		}
	}

	if g.GuildID == "71374620497809408" {
		channel, err := s.UserChannelCreate(g.User.ID)
		if err != nil {
			return
		}
		sendJoinInfo(s, channel)
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

	filters := []*regexp.Regexp{regexp.MustCompile("dick"), regexp.MustCompile("shit"), regexp.MustCompile("fuck"), regexp.MustCompile("penis"), regexp.MustCompile("vagina"), regexp.MustCompile("fag"), regexp.MustCompile("\\brape"), regexp.MustCompile("slut"), regexp.MustCompile("slut"), regexp.MustCompile("hitler"), regexp.MustCompile("\\b(jack)?ass(holes?|lick|wipe)?\\b"), regexp.MustCompile("arse(hole)?"), regexp.MustCompile("bitch"), regexp.MustCompile("whore"), regexp.MustCompile("nigg(er|a)"), regexp.MustCompile("bastard"), regexp.MustCompile("bea?stiality"), regexp.MustCompile("negro"), regexp.MustCompile("retard"), regexp.MustCompile("\\bcum\\b"), regexp.MustCompile("cunt"), regexp.MustCompile("dildo"), regexp.MustCompile("bollocks?"), regexp.MustCompile("\\bwank"), regexp.MustCompile("jizz"), regexp.MustCompile("piss"),}
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
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Messaged removed from <@"+m.Author.ID+">.")
		removeLater(s, rm)
		return
	}

	if !strings.HasPrefix(c, "!") {
		return
	}

	cmdBits := strings.Split(c, " ")

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}

	fmt.Println(cmdBits)

	switch cmdBits[0] {
	case clearCmd.Prefix:
		if !canManageMessage(s, m.Author, d) {
			return
		}
		clearCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, *discordgo.Member, []string))(s, d, m.Message, member, cmdBits)
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
			_, _ = s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		userId := strings.TrimPrefix(strings.TrimSuffix(cmdBits[1], ">"), "<@")
		muteCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, string))(s, d, m.Message, userId)
	case allMuteCmd.Prefix:
		if !admin {
			return
		}
		if !strings.Contains(c, "@") {
			_, _ = s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		userId := strings.TrimPrefix(strings.TrimSuffix(cmdBits[1], ">"), "<@")
		allMuteCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, string))(s, d, m.Message, userId)
	case unMuteAllCmd.Prefix:
		if !admin {
			return
		}
		if !strings.Contains(c, "@") {
			_, _ = s.ChannelMessageSend(d.ID, "Please provide a user to unmute!")
			return
		}
		userId := strings.TrimPrefix(strings.TrimSuffix(cmdBits[1], ">"), "<@")
		unMuteAllCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, string))(s, d, m.Message, userId)
	case broomCmd.Prefix:
		broomCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case rickCmd.Prefix:
		rickCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case vktrsCmd.Prefix:
		vktrsCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case triviaCmd.Prefix:
		triviaCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case topicCmd.Prefix:
		topicCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case joinCmd.Prefix:
		joinCmd.Function.(func(*discordgo.Session, *discordgo.User))(s, m.Author)
	case joinAdmCmd.Prefix:
		joinAdmCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, channel)
	}
}

func doLater(i func()) {
	timer := time.NewTimer(time.Minute * 1)
	<-timer.C
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
	if (uPerms & discordgo.PermissionManageMessages) == discordgo.PermissionManageMessages {
		return true
	}
	return false
}

func clearChannelChat(i int, channel *discordgo.Channel, session *discordgo.Session) {
	fmt.Println("Clearing channel messages...")
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		_, _ = session.ChannelMessageSend(channel.ID, "Could not get messages.")
		_, _ = session.ChannelMessageSend(channel.ID, "```"+err.Error()+"```")
		return
	}
	var todelete []string
	for _, message := range messages {
		todelete = append(todelete, message.ID)
	}
	err = session.ChannelMessagesBulkDelete(channel.ID, todelete)
	if err != nil {
		_, _ = session.ChannelMessageSend(channel.ID, "```"+err.Error()+"```")
		return
	}
	m, err := session.ChannelMessageSend(channel.ID, "Messages removed in channel "+"<#"+channel.ID+">.")
	if err != nil {
		_, _ = session.ChannelMessageSend(channel.ID, "```"+err.Error()+"```")
		return
	}
	removeLater(session, m)
}

func clearUserChat(i int, channel *discordgo.Channel, session *discordgo.Session, id string) {
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		_, _ = session.ChannelMessageSend(channel.ID, "Could not get messages.")
		_, _ = session.ChannelMessageSend(channel.ID, "```"+err.Error()+"```")
		return
	}
	var todelete []string
	for _, message := range messages {
		if message.Author.ID == id {
			todelete = append(todelete, message.ID)
		}
	}
	_ = session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, _ := session.ChannelMessageSend(channel.ID, "Messages removed for user <@"+id+"> in channel "+"<#"+channel.ID+">.")
	removeLater(session, m)
}

func removeLaterBulk(session *discordgo.Session, messages []*discordgo.Message) {
	for _, z := range messages {
		timer := time.NewTimer(time.Second * 5)
		<-timer.C
		_ = session.ChannelMessageDelete(z.ChannelID, z.ID)
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
	<-timer.C
	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func getJson(url string, target interface{}) error {
	stat, body, err := client.Get(nil, url)
	if err != nil || stat != 200 {
		return errors.New("could not obtain json response")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(target)
}

type UnfilteredChannel struct {
	ChannelID string `json:"ID"`
}

type MutedUser struct {
	DiscordID string `json:"ID"`
}

func addToUnfiltered(channelId, guildId string) {
	channel := UnfilteredChannel{ChannelID: channelId}
	if err := db.Write(guildId, channelId, channel); err != nil {
		fmt.Println(err)
	}
}

func removeFromUnfiltered(channelId, guildId string) (err error) {
	if err = db.Delete(guildId, channelId); err != nil {
		return
	}
	return
}

func IsChannelFiltered(channelId, guildId string) (b bool) {
	c := UnfilteredChannel{}
	b = true
	if err := db.Read(guildId, channelId, &c); err != nil {
		return
	}
	b = false
	return
}

func addToMuted(userId, guildId string) {
	user := MutedUser{DiscordID: userId}
	if err := db.Write(guildId, userId, user); err != nil {
		fmt.Println(err)
	}
}

func removeFromMuted(userId, guildId string) (err error) {
	if err = db.Delete(guildId, userId); err != nil {
		return
	}
	return
}

func isMuted(userId, guildId string) bool {
	user := MutedUser{}

	if err := db.Read(guildId, userId, &user); err != nil {
		return false
	}

	return true
}

type ServerInfo struct {
	Name     string
	Launcher string
	Version  string
	Address  string
}

func (server ServerInfo) getDiscordString() string {
	return fmt.Sprintf("%s (Version %s): `%s`", server.Name, server.Version, server.Address)
}

func getServerInfo(s *discordgo.Session) []ServerInfo {
	chann, _ := s.GuildChannels("71374620497809408")

	var serverInfos []ServerInfo

	for _, ch := range chann {
		if ch.ParentID == "360566612220182529" {
			if strings.Contains(ch.Topic, "||") {
				parsedChannels := strings.Split(ch.Topic, "\\||")
				for e := range parsedChannels {
					channelTopicSplit := strings.Split(parsedChannels[e], "|")
					if len(channelTopicSplit) >= 4 {
						serverInfos = append(serverInfos[:], ServerInfo{Name: channelTopicSplit[0], Launcher: channelTopicSplit[1], Version: channelTopicSplit[2], Address: channelTopicSplit[3]})
					}
				}
			} else {
				channel := strings.Split(ch.Topic, "|")
				if len(channel) >= 4 {
					serverInfos = append(serverInfos[:], ServerInfo{Name: channel[0], Launcher: channel[1], Version: channel[2], Address: channel[3]})
				}
			}
		}
	}

	return serverInfos
}

func sendJoinInfo(s *discordgo.Session, mChannel *discordgo.Channel) {
	serverInfo := getServerInfo(s)

	serverInfoStringForm := ""

	for e := range serverInfo {
		serverInfoStringForm += serverInfo[e].getDiscordString() + "\n"
	}

	embed := discordgo.MessageEmbed{
		Title:       "Welcome To Nytro Networks",
		Color:       10181046,
		Description: "Welcome to The Nytro Network official Discord, please read the rules below.",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "About Us", Value: "We are a Modded Minecraft & Ark server network.\n", Inline: false},
			{Name: "Servers", Value: fmt.Sprintf("%s", serverInfoStringForm), Inline: false},
			{Name: "Rules of the Nytro Networks Discord <:Nytro:435179559462109185> :clipboard:", Value: "- Do not bypass the filter.\n- Be nice to everyone.\n- Do not spam, this includes sending the same message in multiple channels, just send it once, someone will see it.\n- No advertising of other servers, publicly or in private messages.\n- Any links that are dangerous or we believe to be suspicious will be removed.", Inline: false},
			{Name: "Support", Value: "If you need help with any server issues, please use <#122793141316091904>.\nIf you're having any issues with our launcher, please use <#435176219542028289>", Inline: false},
			{Name: "Links ", Value: "[Shop](http://shop.nytro.co)\n" +
				"[Website](https://nytro.co)\n", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "shibas are the best cats",
		},
	}
	_, err := s.ChannelMessageSendEmbed(mChannel.ID, &embed)
	if err != nil {
		return
	}
}
