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
	nofilter []string
	triviaRunning = false
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
		fmt.Println("error obtaining account details,", err)
	}

	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	dg.AddHandler(messageCreate)



	BotID = u.ID
	err = dg.Open()
	if err != nil {
		fmt.Println("Could not open discord session: ", err)
	}
	fmt.Println("CatBot is now running.  Press CTRL-C to exit.")
	select {}
}

func forever() {}

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

	if filterChannel(d.ID) {
		for i := 0; i < len(filters); i++ {
			filt := filters[i]
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
		clearCmd.Function.(func(*discordgo.Session, *discordgo.Channel, *discordgo.Message, *discordgo.Member, []string))(s, d, m.Message, member, []string{cmdBits[1], cmdBits[2]})
	case triviaCmd.Prefix:
		triviaCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	case topicCmd.Prefix:
		topicCmd.Function.(func(*discordgo.Session, *discordgo.Channel))(s, d)
	default:
		s.ChannelMessageSend(d.ID, "Unknown command.")
	}
	
}
func doLater(i func()) {
	timer := time.NewTimer(time.Minute * 1)
	<- timer.C
	i()
}

func countChannels(guilds []*discordgo.Guild) (channels int) {
	for i := 0; i < len(guilds); i++ {
		channels = len(guilds[i].Channels) + channels
	}
	return
}

func filterChannel(id string) (b bool) {
	b = true
	//for all the channels without filters,
	for i := 0; i < len(nofilter); i++ {
		//see if nofilter contains the channel id
		if nofilter[i] == id {
			b = false
			return
		}
	}
	return
}

func countUsers(guilds []*discordgo.Guild) (users int) {
	for i := 0; i < len(guilds); i++ {
		users = guilds[i].MemberCount + users
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
	for i := 0; i < len(messages); i++ {
		message := messages[i]
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
	for i := 0; i < len(messages); i++ {
		message := messages[i]
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

func alreadyMuted(id string, channel *discordgo.Channel) (b bool) {
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

func sendLater(s *discordgo.Session, cid string, msg string) {
	timer := time.NewTimer(time.Minute * 1)
	<- timer.C
	s.ChannelMessageSend(cid, msg)
}

//structs
type CatResponse struct {
	URL string `json:"file"`
}

func getJson(url string, target interface{}) error {
	stat, body, err := client.Get(nil, url)
	if err != nil || stat != 200 {
		return errors.New("Could not obtain json response")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(target)

	/*
	resp, err := httpClient.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
	*/
}