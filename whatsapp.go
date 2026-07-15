package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	qrterminal "github.com/mdp/qrterminal/v3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

func GetNewsletterData() (string, error) {
	data, err := GetDetailedNewsletterData()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Total subscribers: %d", data.Total), nil
}

func GetDetailedNewsletterData() (*NewsletterData, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file, relying on environment variables")
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

	data, err := getDetailedSubscriberData(context.Background(), newsletters)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type S3Client struct {
	*s3.Client
}

func NewS3Client(cfg aws.Config) *S3Client {
	return &S3Client{s3.NewFromConfig(cfg)}
}

func (c *S3Client) downloadFromS3(ctx context.Context, bucket, key, dest string) error {
	object, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer object.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, object.Body)
	return err
}

func getDetailedSubscriberData(ctx context.Context, newsletters []NewsletterInfo) (*NewsletterData, error) {
	dbLog := waLog.Stdout("Database", "ERROR", true)

	bucketName := os.Getenv("S3_BUCKET_NAME")
	dbFileName := os.Getenv("DB_FILE_NAME")
	localPath := "/tmp/" + dbFileName

	cfgAws, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	s3Client := NewS3Client(cfgAws)

	fmt.Println("Downloading database from S3...")
	if err := s3Client.downloadFromS3(ctx, bucketName, dbFileName, localPath); err != nil {
		fmt.Printf("Could not download DB (normal for first run): %v\n", err)
	}

	dbString := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&cache=shared",
		localPath,
	)

	container, err := sqlstore.New(ctx, "sqlite", dbString, dbLog)
	if err != nil {
		return nil, fmt.Errorf("sqlstore.New: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetFirstDevice: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))

	if client.Store.ID == nil {
		return nil, fmt.Errorf("no session found in DB — run QR auth locally first, then upload the DB to S3")
	}

	connected := make(chan bool, 1)

	eventHandler := func(evt interface{}) {
		switch v := evt.(type) {
		case *events.QR:
			log.Printf("QR code received, please scan it with your phone:")
			qrterminal.GenerateHalfBlock(v.Codes[0], qrterminal.L, os.Stdout)
		case *events.Connected:
			log.Printf("WhatsApp connected successfully!")
			select {
			case connected <- true:
			default:
			}
		case *events.Disconnected:
			log.Printf("WhatsApp disconnected, reconnecting...")
		case *events.LoggedOut:
			log.Printf("WhatsApp logged out")
		}
	}
	client.AddEventHandler(eventHandler)

	// ✅ Fix 5: single Connect() call
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	log.Printf("Waiting for WhatsApp connection...")
	timeout := time.After(120 * time.Second)

	select {
	case <-connected:
		log.Printf("WhatsApp connected! Pausing for 5 seconds to stabilize...")
		time.Sleep(5 * time.Second)
	case <-timeout:
		return nil, fmt.Errorf("timeout waiting for WhatsApp connection")
	}

	log.Printf("WhatsApp connection established. Fetching newsletter data...")

	var updatedNewsletters []NewsletterInfo
	var totalSubscribers int

	for _, newsletter := range newsletters {
		jid, err := types.ParseJID(newsletter.JID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JID for %s: %v", newsletter.Name, err)
		}

		info, err := client.GetNewsletterInfo(ctx, jid)
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

	log.Printf("Data fetched. Keeping session alive for 30 seconds to stabilize...")
	time.Sleep(30 * time.Second)
	log.Printf("Disconnecting now.")
	client.Disconnect()

	return &NewsletterData{
		Total:       totalSubscribers,
		Newsletters: updatedNewsletters,
	}, nil
}
