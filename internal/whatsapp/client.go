package whatsapp

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	qrterminal "github.com/mdp/qrterminal/v3"

	_ "github.com/mattn/go-sqlite3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type NewsletterInfo struct {
	Name        string
	JID         string
	Subscribers int
}

type NewsletterData struct {
	Total       int
	Newsletters []NewsletterInfo
}

func GetNewsletterData() (string, error) {
	data, err := GetDetailedNewsletterData()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Total subscribers: %d", data.Total), nil
}

func GetDetailedNewsletterData() (*NewsletterData, error) {
	if err := godotenv.Load("/root/.env"); err != nil {
		// It's okay if .env file doesn't exist, environment variables can be set by Docker
		fmt.Println("Warning: Could not load .env file, relying on environment variables")
	}

	jidStrNumber := os.Getenv("NUMBER_NEWSLETTERS")
	if jidStrNumber == "" {
		return nil, fmt.Errorf("NUMBER_NEWSLETTERS environment variable not set")
	}

	numNewsletters, err := strconv.Atoi(jidStrNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid NUMBER_NEWSLETTERS value: %v", err)
	}

	var newsletters []NewsletterInfo
	for i := 1; i <= numNewsletters; i++ {
		jidStr := os.Getenv(fmt.Sprintf("NEWSLETTER_JID%d", i))
		if jidStr == "" {
			return nil, fmt.Errorf("NEWSLETTER_JID%d environment variable not set", i)
		}

		nameStr := os.Getenv(fmt.Sprintf("NEWSLETTER_NAME%d", i))
		if nameStr == "" {
			nameStr = fmt.Sprintf("Newsletter %d", i)
		}

		newsletters = append(newsletters, NewsletterInfo{
			Name: nameStr,
			JID:  jidStr,
		})
	}

	data, err := getDetailedSubscriberData(newsletters)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getDetailedSubscriberData(newsletters []NewsletterInfo) (*NewsletterData, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)

	dbPath := "file:/root/db/session.db?_foreign_keys=on"

	container, err := sqlstore.New(context.Background(), "sqlite3", dbPath, dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	connected := make(chan bool, 1)
	reconnecting := make(chan bool, 1)

	eventHandler := func(evt interface{}) {
		switch v := evt.(type) {
		case *events.QR:
			fmt.Println("QR code received, please scan it with your phone:")
			qrterminal.GenerateHalfBlock(v.Codes[0], qrterminal.L, os.Stdout)
		case *events.Connected:
			fmt.Println("WhatsApp connected successfully!")
			select {
			case connected <- true:
			default:
			}
		case *events.Disconnected:
			fmt.Println("WhatsApp disconnected, reconnecting...")
			select {
			case reconnecting <- true:
			default:
			}
		case *events.LoggedOut:
			fmt.Println("WhatsApp logged out")
		}
	}
	client.AddEventHandler(eventHandler)

	err = client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Waiting for WhatsApp connection and synchronization...")
	maxWaitTime := 120 * time.Second
	timeout := time.After(maxWaitTime)
	connectionStable := false

	for !connectionStable {
		select {
		case <-connected:
			fmt.Println("Connected! Waiting for synchronization to complete...")
			stabilityCheck := time.After(10 * time.Second)
			stable := true

		stabilityLoop:
			for stable {
				select {
				case <-reconnecting:
					fmt.Println("Reconnection detected, waiting for stability...")
					stable = false
					break stabilityLoop
				case <-stabilityCheck:
					connectionStable = true
					break stabilityLoop
				case <-timeout:
					return nil, fmt.Errorf("timeout waiting for stable connection")
				}
			}

		case <-reconnecting:
			continue
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for WhatsApp connection")
		}
	}

	fmt.Println("WhatsApp connection is stable. Fetching newsletter data...")
	time.Sleep(2 * time.Second)

	var updatedNewsletters []NewsletterInfo
	var totalSubscribers int

	for _, newsletter := range newsletters {
		jid, err := types.ParseJID(newsletter.JID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JID for %s: %v", newsletter.Name, err)
		}

		info, err := client.GetNewsletterInfo(jid)
		if err != nil {
			return nil, fmt.Errorf("failed to get newsletter info for %s: %v", newsletter.Name, err)
		}

		subscribers := int(info.ThreadMeta.SubscriberCount)
		totalSubscribers += subscribers

		updatedNewsletters = append(updatedNewsletters, NewsletterInfo{
			Name:        newsletter.Name,
			JID:         newsletter.JID,
			Subscribers: subscribers,
		})
	}

	return &NewsletterData{
		Total:       totalSubscribers,
		Newsletters: updatedNewsletters,
	}, nil
}
