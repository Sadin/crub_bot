package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jonas747/dca"

	"github.com/bwmarrin/discordgo"

	"github.com/jackc/pgx/v4"
)

// system vairables from command line params
var (
	Token  string
	buffer = make([][]byte, 0)
)

var conn *pgx.Conn

var guild_list []string



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
	}

	fmt.Println("success.")

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)
	// register message create event handler
	dg.AddHandler(messageCreate)

	// open websocket conection to discord
	fmt.Print("connecting to discord...")
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// wait for CTRL-C or term signal
	fmt.Println("Punx operational. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// close
	dg.Close()
}

// This function will be called for every server crub has joined
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		fmt.Println("error with guild" + event.Guild.ID)
		return
	}
	guild_list = append(guild_list,event.Guild.Name)
	fmt.Println(event.Guild.Name)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages sent by bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// print messages to terminal
	message := fmt.Sprintf("%s,%s,\"%s\"", m.Author, m.ChannelID, m.Content)
	fmt.Println(writeLog(message, "logs.csv"))

	// debug info command
	if strings.HasPrefix(strings.ToLower(m.Content), "!debug") {
		// get data
		debug, err := s.Guild(m.GuildID)
		if err != nil {
			// error fetching guild debug
			return
		}
		// print debug header
		res := fmt.Sprintf("**DEBUG INFO** \n **ID:** `%s` \n **Name:** `%s` \n **Region:** `%s` \n", debug.ID, debug.Name, debug.Region)
		// loop through guild name slice, append to debug response
		for guild_list_item := 0; guild_list_item < len(guild_list); guild_list_item++ {
			// header
			if guild_list_item == 0 {
				res = res + fmt.Sprintf("\n**CONNECTED SERVERS** \n")
			}
			// append guild name to debug response
			res = res + fmt.Sprintf("`%s`\n",guild_list[guild_list_item])
		}
		// send res to user
		s.ChannelMessageSend(m.ChannelID, res)
	}

	// call bot into voice
	if strings.HasPrefix(m.Content, "!hey") || strings.HasPrefix(m.Content, "!hi") {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			msg := fmt.Sprintf("%s", err)
			writeLog(msg, "logs.csv")
			return
		}
		fmt.Println(m.ChannelID)
		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			msg := fmt.Sprintf("%s", err)
			writeLog(msg, "logs.csv")
			return
		}
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = joinVoice(s, g.ID, vs.ChannelID, m)
				if err != nil {
					msg := fmt.Sprintf("%s", err)
					writeLog(msg, "logs.csv")
				}
				return
			}
		}
	}

	// dismiss bot
	if strings.HasPrefix(m.Content, "!bye") {
		// store session VoiceConnections
		connections := s.VoiceConnections
		// key for VoiceConnection using GuildID issues by command
		vc := connections[m.GuildID]
		// play ending sound if suffix is sent
		if strings.HasSuffix(strings.ToLower(m.Content), "xqcl") {
			vc.Speaking(true)
			err := playSound("./audio_files/for_the_day_FULL.dca", vc)
			if err != nil {
				msg := fmt.Sprintf("%s", err)
				writeLog(msg, "logs.csv")
			} else {
				for _, buff := range buffer {
					vc.OpusSend <- buff
				}
			}
			vc.Speaking(false)
		}

		// close
		err := vc.Disconnect()
		if err != nil {
			msg := fmt.Sprintf("%s", err)
			writeLog(msg, "logs.csv")
		}
		return
	}

}

func joinVoice(s *discordgo.Session, g string, c string, m *discordgo.MessageCreate) (err error) {
	// join voice channel

	// log
	vcdebug := fmt.Sprintf("%s,%s,Guild: %s | Channel: %s", m.Author, m.ChannelID, g, "725730150989168730")
	fmt.Println(writeLog(vcdebug, "logs.csv"))
	// join
	vc, err := s.ChannelVoiceJoin(g, c, false, false)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	vc.Speaking(true)

	err = playSound("./audio_files/crubb.dca", vc)
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		writeLog(msg, "logs.csv")
		vc.Speaking(false)
		return
	}
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	vc.Speaking(false)

	return
}

func aroundWorld(victim string, s *discordgo.Session) {
	// hahaha this is it

}

func isConnected(user string, s *discordgo.Session) bool {
	// check if user is in voice
	return false
}

// Audio Functions -  load file, play audio

func playSound(filepath string, vc *discordgo.VoiceConnection) error {

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	decoder := dca.NewDecoder(file)

	for {
		frame, err := decoder.OpusFrame()
		if err != nil {
			if err != io.EOF {
				msg := fmt.Sprintf("Error reading from dca file (1): %s", err)
				writeLog(msg, "logs.csv")
			}

			break
		}

		// send frame to VoiceConnection
		select {
		case vc.OpusSend <- frame:
		case <-time.After(time.Second):

			return err
		}
	}
	return err
}

// Utility Functions - Logging, reading configs, etc

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeLog(text string, path string) string {
	// generic function to log text to specified filename, returns string for console
	if fileExists(path) == false {
		f, err := os.Create(path)
		if err != nil {
			e := fmt.Sprintf("Error creating file: %s", path)
			fmt.Println(e)
			f.Close()
			return e
		}
		f.Close()
		fmt.Println(writeLog(text, path))
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		e := fmt.Sprintf("Error opening file: %s", path)
		fmt.Println(e)
		return e
	}

	l, err := fmt.Fprintln(f, text)
	if err != nil {
		e := fmt.Sprintf("Error writing file: %s", path)
		fmt.Println(e)
		f.Close()
		return e
	}
	fmt.Println(l, "bytes written successfully to logs")
	err = f.Close()
	if err != nil {
		e := fmt.Sprintf("Error closing file: %s", path)
		fmt.Println(e)
		return e
	}
	return text
}

func log_chat(m *discordgo.MessageCreate) error {
	fmt.Println(m.Author, m.ChannelID, m.Content)
	_, err := conn.Exec(context.Background(), "insert into crub.messages(message_id,user_name,fk_user_id,message_text) values(nextval('crub.message_id_seq'),$1,$2,$3)", m.Author, m.ChannelID, m.Content)
	return err
}
