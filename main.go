package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// system vairables from command line params
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	fmt.Println("Punx")

	// launch discord session
	fmt.Print("creating new discord session...")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	} else {
		fmt.Println("success.")
	}

	// register message create event handler
	dg.AddHandler(messageCreate)

	// open websocket conection to discord
	fmt.Print("connecting to discord...")
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	} else {
		fmt.Println("success!")
	}

	// wait for CTRL-C or term signal
	fmt.Println("Punx operational. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// close
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages sent by bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// print messages to terminal
	fmt.Printf("%s %s \"%s\" \n", m.Author, m.ChannelID, m.Content)

	if m.Content == "!Punx" {
		s.ChannelMessageSend(m.ChannelID, "Go fuck yourself, im not ready")
	}
}
