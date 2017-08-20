package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"fmt"
	"math/rand"
	"html"
)

func info(s *discordgo.Session, d *discordgo.Channel) {
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
}

func removeFilter(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message) {
	if filterChannel(d.ID) == false {
		e, _ := s.ChannelMessageSend(d.ID, "Channel already unfiltered.")
		removeLaterBulk(s, []*discordgo.Message{e, m})
	} else {
		nofilter = append(nofilter, d.ID)
		e, _ := s.ChannelMessageSend(d.ID, "Channel is no longer filtered.")
		removeLaterBulk(s, []*discordgo.Message{e, m})
	}
}

func enableFilter(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message) {
	if filterChannel(d.ID) == false {
		toremove := -1
		for i := range nofilter {
			if nofilter[i] == d.ID {
				toremove = i
			}
		}
		nofilter = append(nofilter[:toremove], nofilter[toremove+1:]...)
		e, _ := s.ChannelMessageSend(d.ID, "Channel is now filtered.")
		removeLaterBulk(s, []*discordgo.Message{e, m})
	} else {
		e, _ := s.ChannelMessageSend(d.ID, "Channel is already filtered.")
		removeLaterBulk(s, []*discordgo.Message{e, m})
	}
}

func catbot(s *discordgo.Session, d *discordgo.Channel) {
	s.ChannelMessageSend(d.ID, "Meow meow beep boop! I am catbot 2.0!")
}

func mute(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message, user_id string) {
	if !alreadyMutedInChannel(user_id, d) {
		s.ChannelPermissionSet(d.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
		rm, _ := s.ChannelMessageSend(d.ID, "Muted user <@" + user_id + ">!")
		fmt.Println(m.Author.Username + " muted " + user_id)
		removeLaterBulk(s, []*discordgo.Message{rm, m,})
	} else {
		rm, _ := s.ChannelMessageSend(m.ChannelID, "User already muted!")
		removeLaterBulk(s, []*discordgo.Message{rm, m,})
	}
	fmt.Println(m.Author.Username + " muted " + user_id + " in all channels.")
	addToMuted(user_id, d.GuildID)
}

func allMute(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message, user_id string) {
	channels, _ := s.GuildChannels(d.GuildID)
	for i := 0; i < len(channels); i++ {
		channel := channels[i]
		if !alreadyMutedInChannel(user_id, channel) {
			s.ChannelPermissionSet(channel.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
		}
	}
	rm, _ := s.ChannelMessageSend(d.ID, "Muted user <@" + user_id + "> in all channels!")
	removeLaterBulk(s, []*discordgo.Message{rm, m,})
	fmt.Println(m.Author.Username + " muted " + user_id + " in all channels.")
	addToMuted(user_id, d.GuildID)
}

func donationHelp(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message) {
	s.ChannelMessageSend(d.ID,"If you don't have a rank or perk you purchased please make a forum post here: http://kkmc.info/2du3U2l")
	removeLater(s, m)
}

func cat(s *discordgo.Session, d *discordgo.Channel) {
	j := CatResponse{}
	getJson("http://random.cat/meow", &j)
	//s.ChannelMessageSend(d.ID, j.URL)
	s.ChannelMessageSendEmbed(d.ID, &discordgo.MessageEmbed{Color: 10181046, Image: &discordgo.MessageEmbedImage{URL:j.URL}, Title: "Cat", URL: j.URL})
}

func snek(s *discordgo.Session, d *discordgo.Channel) {
	j := CatResponse{}
	getJson("http://fur.im/snek/snek.php", &j)
	//s.ChannelMessageSend(d.ID, j.URL)
	s.ChannelMessageSendEmbed(d.ID, &discordgo.MessageEmbed{Color: 10181046, Image: &discordgo.MessageEmbedImage{URL:j.URL}, Title: "Snek", URL: j.URL})
}

func broom(s *discordgo.Session, d *discordgo.Channel) {
	s.ChannelMessageSend(d.ID, "https://youtu.be/sSPIMgtcQnU")
}

func rick(s *discordgo.Session, d *discordgo.Channel) {
	s.ChannelMessageSend(d.ID, "http://kkmc.info/1LWYru2")
}

func vktrs(s *discordgo.Session, d *discordgo.Channel) {
	s.ChannelMessageSend(d.ID, "http://kkmc.info/hRdfdSD")
}

func clear(s *discordgo.Session, d *discordgo.Channel, m *discordgo.Message, member *discordgo.Member, args []string) {
	fmt.Println("clearing messages...")
	if len(args) == 0 {
		s.ChannelMessageSend(d.ID, "Invalid parameters")
		fmt.Println("Invalid clear paramters...")
		return
	} else if len(args) == 2 {
		fmt.Println("clearing messages from " + d.Name + " for user " + member.User.Username)
		if i, err := strconv.ParseInt(args[2], 10, 64); err == nil {
			clearUserChat(int(i), d, s, args[1])
			removeLater(s, m)
			return
		}
	} else if len(args) == 1 {
		fmt.Println("clearing " + args[1] + " messages from " + d.Name + " for user " + member.User.Username)
		if i, err := strconv.ParseInt(args[1], 10, 64); err == nil {
			clearChannelChat(int(i), d, s)
			removeLater(s, m)
			return
		}
	}
}

func triviaExec(s *discordgo.Session, d *discordgo.Channel) {
	if triviaRunning {
		s.ChannelMessageSend(d.ID, "Trivia already running.")
	} else {
		fmt.Println("Getting trivia")
		if question, err := trivia.Getter.GetTrivia(1); err == nil {
			triviaRunning = true
			a := append(question.Results[0].IncorrectAnswer, question.Results[0].CorrectAnswer)
			for i := range a {
				j := rand.Intn(i + 1)
				a[i], a[j] = a[j], a[i]
			}
			embedanswers := []*discordgo.MessageEmbedField{}
			if len(a) == 2 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "Category", Value: question.Results[0].Category, Inline: false},
					{Name: "Difficulty", Value: question.Results[0].Difficulty, Inline: false},
					{Name: "A", Value: html.UnescapeString(a[0]), Inline: true},
					{Name: "B", Value: html.UnescapeString(a[1]), Inline: true},
				}
			} else if len(a) == 4 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "Category", Value: question.Results[0].Category, Inline: false},
					{Name: "Difficulty", Value: question.Results[0].Difficulty, Inline: false},
					{Name: "A", Value: html.UnescapeString(a[0]), Inline: true},
					{Name: "B", Value: html.UnescapeString(a[1]), Inline: true},
					{Name: "C", Value: html.UnescapeString(a[2]), Inline: true},
					{Name: "D", Value: html.UnescapeString(a[3]), Inline: true},
				}
			}
			embed := discordgo.MessageEmbed{
				Title:       "Trivia",
				Color:       10181046,
				Description: html.UnescapeString(question.Results[0].Question),
				URL:         "https://opentdb.com/",
				Fields:      embedanswers,
			}
			_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
			if err != nil {
				s.ChannelMessageSend(d.ID, formatError(err))
			}
			doLater(func (){
				s.ChannelMessageSend(d.ID, "The correct answer was: " + html.UnescapeString(question.Results[0].CorrectAnswer))
				triviaRunning = false
			})
		} else if err != nil {
			s.ChannelMessageSend(d.ID, formatError(err))
			fmt.Errorf("Could not get trivia", err)
		}
	}
}

func topic(s *discordgo.Session, d *discordgo.Channel) {
	s.ChannelMessageSendEmbed(d.ID, &discordgo.MessageEmbed{Description: d.Topic, Title: d.Name, Color: 10181046,})
}

func help(s *discordgo.Session, d *discordgo.Channel, user *discordgo.User, admin bool) {
	fmt.Println("cb help executed")
	embedElements := []*discordgo.MessageEmbedField{}
	for _, cmd := range cmds {
		if cmd.AdminReq && !admin || cmd.Description == "Secret." {
			continue
		} else {
			embedElements = append(embedElements, &discordgo.MessageEmbedField{Name:cmd.Prefix, Inline: false, Value:cmd.Description})
		}
	}

	channel, err := s.UserChannelCreate(user.ID)
	if err != nil {
		fmt.Errorf("Could not create private channel", err)
	}
	s.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{Title:"Catbot Help", Fields:embedElements, Color:10181046})
}