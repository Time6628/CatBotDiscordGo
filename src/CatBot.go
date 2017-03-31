package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"flag"
	"time"
	"regexp"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var BotID string

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

func forever() {

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	d, _ := s.Channel(m.ChannelID)
	g, _ := s.Guild(d.GuildID)
	member, _ := s.GuildMember(g.ID, m.Author.ID)
	roles := member.Roles
	//groles := g.Roles

	/*
	defperm := 0
	for i := 0; i < len(groles); i++ {
		ro := groles[i]
		if ro.Position == 0 {
			defperm = ro.Permissions
		}
	}*/
	//fmt.Println("User permissions: ", adminperm)
	//fmt.Println("Admin perm:", admin)

	c := strings.ToLower(m.Content)

	filters := [...]*regexp.Regexp{
		regexp.MustCompile("dick"),
		regexp.MustCompile("fuck"),
		regexp.MustCompile("penis"),
		regexp.MustCompile("vagina"),
		regexp.MustCompile("fag"),
		regexp.MustCompile("\\brape"),
		regexp.MustCompile("slut"),
		regexp.MustCompile("slut"),
		regexp.MustCompile("hitler"),
		regexp.MustCompile("\\b(jack)?ass(holes?|lick|wipe)?\\b"),
		regexp.MustCompile("arse(hole)?"),
		regexp.MustCompile("bitch"),
		regexp.MustCompile("whore"),
		regexp.MustCompile("nigg(er|a)"),
		regexp.MustCompile("bastard"),
		regexp.MustCompile("bea?stiality"),
		regexp.MustCompile("negro"),
		regexp.MustCompile("retard"),
		regexp.MustCompile("\\bcum\\b"),
		regexp.MustCompile("cunt"),
		regexp.MustCompile("dildo"),
		regexp.MustCompile("bollocks?"),
		regexp.MustCompile("\\bwank"),
		regexp.MustCompile("jizz"),
		regexp.MustCompile("piss"),
	}

	filter := false
	for i := 0; i < len(filters); i++ {
		filt := filters[i]
		if filt.MatchString(c) {
			filter = true
		}
	}


	admin := false
	for i := 0; i < len(roles); i++ {
		role, _ := s.State.Role(g.ID, roles[i])
		if (role.Permissions & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator {
			admin = true
		}
	}

	if filter && !admin {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Messaged removed from <@" + m.Author.ID + ">.")
		removeLater(s, rm)
		return
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
		arg := strings.Split(cc, ":")
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
	}

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
	//fmt.Println("user perms: ", permissions)
	//fmt.Println("send message perms: ", discordgo.PermissionSendMessages)
	return //if the user's permission in that channel are already lower then SendMessages, they are muted.
}

func removeLater(s *discordgo.Session, m *discordgo.Message) {
	timer := time.NewTimer(time.Second * 5)
	<- timer.C
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}
