package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chrisb92/tc-build-bot/bot/config"
	"github.com/spf13/viper"
)

func init() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Unable to read config: ", err)
		return
	}

	if err := viper.Unmarshal(&conf); err != nil {
		fmt.Println("Unable to decode config: ", err)
		return
	}
}

var discord *discordgo.Session
var conf config.Configuration

func main() {

	if conf.Token == "" {
		fmt.Println("No token provided. Please check your config.json for a token")
		return
	}

	dg, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	discord = dg

	http.HandleFunc("/build", httpBuild)
	go func() {
		fmt.Println("Setting up HTTP server to listen for web hook POST")
		if err := http.ListenAndServe(":5436", nil); err != nil {
			log.Fatal(err)
		}
	}()

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Build watchin'")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.ToLower(m.Content) == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

func httpBuild(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+conf.AuthToken {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	if r.URL.Path != "/build" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		webHookContent := TeamCityBuild{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&webHookContent)
		if err != nil {
			fmt.Println("Failed to decode POST to struct: ", err)
			return
		}

		var embeds []*discordgo.MessageEmbed

		for _, attachment := range webHookContent.Attachments {

			var embedColour int
			if attachment.Color == "danger" {
				embedColour = 14375008
			} else if attachment.Color == "good" {
				embedColour = 5875817
			} else {
				embedColour = 14408667
			}

			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: webHookContent.Username,
				},
				Color: embedColour,
				Title: attachment.Fallback,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: webHookContent.IconURL,
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}

			for _, field := range attachment.Fields {

				fieldValue := field.Value

				if strings.Contains(field.Value, "://") {
					fieldSplit := strings.SplitAfter(fieldValue, "]")
					fieldSplit[1] = strings.Replace(fieldSplit[1], " ", "%20", -1)

					fieldValue = strings.Join(fieldSplit, "")
				}

				field := &discordgo.MessageEmbedField{
					Name:   field.Title,
					Value:  fieldValue,
					Inline: field.Short || false,
				}

				embed.Fields = append(embed.Fields, field)
			}

			embeds = append(embeds, embed)
		}

		for _, embed := range embeds {
			discord.ChannelMessageSendEmbed(conf.MainChannelID, embed)
		}

	default:
		fmt.Fprintln(w, "Only POST is supported")
	}
}

// TeamCityBuild struct for WebHook template
type TeamCityBuild struct {
	Username    string `json:"username"`
	IconURL     string `json:"icon_url"`
	Attachments []struct {
		Title    string `json:"title"`
		Fallback string `json:"fallback"`
		Color    string `json:"color"`
		Fields   []struct {
			Title string `json:"title"`
			Value string `json:"value"`
			Short bool   `json:"short,omitempty"`
		} `json:"fields"`
	} `json:"attachments"`
}
