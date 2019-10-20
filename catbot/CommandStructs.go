package main

type CommandBase struct {
	Prefix      string
	Function    interface{}
	Description string
	Usage       string
	AdminReq    bool
}

var (
	removeFilterCmd = CommandBase{Prefix: "!removefilter", Function: removeFilter, AdminReq: true,  Description: "Removes the filter from a channel.",                    Usage: "!removefilter"}
	enableFilterCmd = CommandBase{Prefix: "!enablefilter", Function: enableFilter, AdminReq: true,  Description: "Removes the filter from a channel.",                    Usage: "!removefilter"}
	infoCmd         = CommandBase{Prefix: "!info",         Function: info,         AdminReq: false, Description: "Catbot stats.",                                         Usage: "!info"}
	catbotCmd       = CommandBase{Prefix: "!catbot",       Function: catbot,       AdminReq: false, Description: "Catbot beep boops.",                                    Usage: "!catbot"}
	muteCmd         = CommandBase{Prefix: "!mute",         Function: mute,         AdminReq: true,  Description: "Mute a user.",                                          Usage: "!mute @User"}
	allMuteCmd      = CommandBase{Prefix: "!allmute",      Function: allMute,      AdminReq: true,  Description: "Mute a user in all of the current guild's channels.",   Usage: "!allmute <@User>"}
	unMuteAllCmd    = CommandBase{Prefix: "!unmuteall",    Function: unMuteAll,    AdminReq: true,  Description: "Unmute a user in all of the current guild's channels.", Usage: "!allmute <@User>"}
	broomCmd        = CommandBase{Prefix: "!broom",        Function: broom,        AdminReq: false, Description: "Stop being a broom.",                                   Usage: "!broom"}
	rickCmd         = CommandBase{Prefix: "!rick",         Function: rick,         AdminReq: false, Description: "Secret.",                                               Usage: "!rick"}
	vktrsCmd        = CommandBase{Prefix: "!vktrs",        Function: vktrs,        AdminReq: false, Description: "Secret.",                                               Usage: "!vktrs"}
	clearCmd        = CommandBase{Prefix: "!clear",        Function: clear,        AdminReq: true,  Description: "Clears messages from a channel or user.",               Usage: "!clear <count> [user]"}
	triviaCmd       = CommandBase{Prefix: "!trivia",       Function: triviaExec,   AdminReq: true,  Description: "Play some trivia.",                                     Usage: "!trivia"}
	topicCmd        = CommandBase{Prefix: "!topic",        Function: topic,        AdminReq: false, Description: "Gets the channel topic.",                               Usage: "!topic"}
	helpCmd         = CommandBase{Prefix: "!help",         Function: help,         AdminReq: false, Description: "See all commands in catbot.",                           Usage: "!help"}
	joinCmd         = CommandBase{Prefix: "!join",         Function: join,         AdminReq: false, Description: "See the guild joining message.",                        Usage: "!join"}
	joinAdmCmd      = CommandBase{Prefix: "!joinadm",      Function: joinAdm,      AdminReq: true,  Description: "See the guild joining message.",                      Usage: "!joinadm"}
	cmds            = []CommandBase{removeFilterCmd, enableFilterCmd, infoCmd, catbotCmd, muteCmd, allMuteCmd, unMuteAllCmd, broomCmd, rickCmd, vktrsCmd, clearCmd, triviaCmd, topicCmd, joinCmd}
)