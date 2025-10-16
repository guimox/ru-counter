package whatsapp

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func GetNewsletterData() (string, error) {
	_ = godotenv.Load()

	jidStr := os.Getenv("NEWSLETTER_JID")
	if jidStr == "" {
		return "", fmt.Errorf("NEWSLETTER_JID environment variable not set")
	}

	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:db/session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return "", err
	}
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return "", err
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	err = client.Connect()
	if err != nil {
		return "", err
	}
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return "", err
	}
	info, err := client.GetNewsletterInfo(jid)
	if err != nil {
		return "", err
	}
	result := fmt.Sprintf(
		"Newsletter name: %s\nSubscriber count: %d\n",
		info.ThreadMeta.Name, info.ThreadMeta.SubscriberCount)
	return result, nil
}
