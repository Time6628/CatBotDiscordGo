package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"flag"
	"time"
	"regexp"
	"encoding/json"
	"strconv"
	"github.com/valyala/fasthttp"
	"errors"
	"bytes"
	"github.com/Time6628/OpenTDB-Go"
	"math/rand"
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

	d, _ := s.Channel(m.ChannelID)
	if d.IsPrivate {
		return
	}

	g, _ := s.Guild(d.GuildID)
	member, _ := s.GuildMember(g.ID, m.Author.ID)
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
	} else if strings.HasPrefix(c, "!removefilter") {
		if filterChannel(d.ID) == false {
			s.ChannelMessageSend(d.ID, "Channel already unfiltered.")
		} else {
			nofilter = append(nofilter, d.ID)
			s.ChannelMessageSend(d.ID, "Channel is no longer filtered.")
		}
	} else if strings.HasPrefix(c, "!enablefilter") {
		if filterChannel(d.ID) == false {
			toremove := -1
			for i := range nofilter {
				if nofilter[i] == d.ID {
					toremove = i
				}
			}
			nofilter = append(nofilter[:toremove], nofilter[toremove+1:]...)
			s.ChannelMessageSend(d.ID, "Channel is now filtered.")
		} else {
			s.ChannelMessageSend(d.ID, "Channel is already filtered.")
		}
	} else if strings.HasPrefix(c, "!catbot") {
		s.ChannelMessageSend(m.ChannelID, "Meow meow beep boop! I am catbot 2.0!")
		return
	} else if strings.HasPrefix(c, "!mute") && admin {
		cc := strings.TrimPrefix(c, "!mute ")
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		arg := strings.Split(cc, " ")
		fmt.Println(arg[0])
		fmt.Println(cc)
		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
		if !alreadyMuted(user_id, d) {
			s.ChannelPermissionSet(d.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
			rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user " + arg[0] + "!")
			fmt.Println(m.Author.Username + " muted " + user_id)
			b := []*discordgo.Message{rm, m.Message,}
			removeLaterBulk(s, b)
		} else {
			rm, _ := s.ChannelMessageSend(m.ChannelID, "User already muted!")
			b := []*discordgo.Message{rm, m.Message,}
			removeLaterBulk(s, b)
		}
	} else if strings.HasPrefix(c, "!allmute") && admin {
		cc := strings.TrimPrefix(c, "!allmute ")
		arg := strings.Split(cc, " ")
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
		channels := g.Channels
		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			if !alreadyMuted(user_id, channel) {
				s.ChannelPermissionSet(channel.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
			}
		}
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user " + arg[0] + " in all channels!")
		b := []*discordgo.Message{rm, m.Message,}
		removeLaterBulk(s, b)
		fmt.Println(m.Author.Username + " muted " + user_id + " in all channels.")
	} else if strings.HasPrefix(c, "!donationhelp") {
		s.ChannelMessageSend(m.ChannelID,"If you don't have a rank or perk you purchased please make a forum post here: http://kkmc.info/2du3U2l")
		removeLater(s, m.Message)
	} else if strings.HasPrefix(c, "!cat") {
		fmt.Println(time.Now())
		j := CatResponse{}
		cc := strings.TrimPrefix(c, "!cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
			fmt.Println(time.Now())
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://random.cat/meow", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
			fmt.Println(time.Now())
		}
	} else if strings.HasPrefix(c, "!snek") {
		j := CatResponse{}
		cc := strings.TrimPrefix(c, "!snek ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://fur.im/snek/snek.php", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://fur.im/snek/snek.php", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, "!broom") || strings.HasPrefix(c, "!dontbeabroom") {
		s.ChannelMessageSend(d.ID, "https://youtu.be/sSPIMgtcQnU")
	} else if strings.HasPrefix(c, "!rick") {
		s.ChannelMessageSend(d.ID, "http://kkmc.info/1LWYru2")
	} else if strings.HasPrefix(c, "!clear") {
		if len(c) < 7  || !canManageMessage(s, m.Author, d) {
			return
		}
		fmt.Println("clearing messages...")
		args := strings.Split(strings.Replace(c, "!clear ", "", -1), " ")
		if len(args) == 0 {
			s.ChannelMessageSend(d.ID, "Invalid parameters")
			fmt.Println("Invalid clear paramters...")
			return
		} else if len(args) == 2 {
			fmt.Println("clearing user messages...")
			if i, err := strconv.ParseInt(args[1], 10, 64); err != nil {
				clearUserChat(int(i), d, s, args[0])
				removeLater(s, m.Message)
				return
			}
		} else if len(args) == 1 {
			fmt.Println("clearing messages...")
			if i, err := strconv.ParseInt(args[0], 10, 64); err != nil {
				clearChannelChat(int(i), d, s)
				removeLater(s, m.Message)
				return
			}
		}
	} else if strings.HasPrefix(c, "!info") {
		fmt.Println("Sending info...")
		embed := discordgo.MessageEmbed{
			Title: "Info",
			Color: 10181046,
			Description: "A rewrite of KookyKraftMC discord bot, written in Go.",
			URL: "https://github.com/Time6628/CatBotDiscordGo",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Servers", Value: strconv.Itoa(len(s.State.Guilds)), Inline: true},
				{Name: "Users", Value: strconv.Itoa(countUsers(s.State.Guilds)), Inline: true},
				{Name: "Channels", Value: strconv.Itoa(countChannels(s.State.Guilds)), Inline: true},
			},
		}
		_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
		if err != nil {
			s.ChannelMessageSend(d.ID, formatError(err))
		}
	} else if strings.HasPrefix(c, "!trivia") {
		fmt.Println("Getting trivia")
		if question, err := trivia.Getter.GetTrivia(1); err == nil {
			a := append(question.Results[0].IncorrectAnswer, question.Results[0].CorrectAnswer)
			for i := range a {
				j := rand.Intn(i + 1)
				a[i], a[j] = a[j], a[i]
			}
			embedanswers := []*discordgo.MessageEmbedField{}
			if len(a) == 2 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "A", Value: a[0], Inline: true},
					{Name: "B", Value: a[1], Inline: true},
				}
			} else if len(a) == 4 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "A", Value: a[0], Inline: true},
					{Name: "B", Value: a[1], Inline: true},
					{Name: "C", Value: a[2], Inline: true},
					{Name: "D", Value: a[3], Inline: true},
				}
			}
			embed := discordgo.MessageEmbed{
				Title: "Trivia",
				Color: 10181046,
				Description: question.Results[0].Question,
				URL: "https://opentdb.com/",
				Fields: embedanswers,
			}
			_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
			if err != nil {
				s.ChannelMessageSend(d.ID, formatError(err))
			}
			sendLater(s, d.ID, "The correct answer was: " + question.Results[0].CorrectAnswer)
		} else if err != nil {
			s.ChannelMessageSend(d.ID, formatError(err))
		}
	}
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
	messages, err := session.ChannelMessages(channel.ID, i, "", "")
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not get messages.")
		return
	}
	todelete := []string{}
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		todelete = append(todelete, message.ID)
	}
	session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, _ := session.ChannelMessageSend(channel.ID, "Messages removed in channel " + channel.Name)
	removeLater(session, m)
}

func clearUserChat(i int, channel *discordgo.Channel, session *discordgo.Session, id string) {
	messages, err := session.ChannelMessages(channel.ID, i, "", "")
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