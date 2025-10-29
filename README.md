# RU Counter

Automated WhatsApp newsletter subscriber counter for the [ru-menu](https://github.com/guimox/ru-menu) project. This service automatically fetches subscriber counts from WhatsApp newsletters and updates GitHub repository statistics.

### Overview

RU Counter is a microservice that connects to WhatsApp Business API to retrieve real-time subscriber counts from university restaurant menu newsletters. It automatically updates the main repository description and README with current Daily Active User (DAU) statistics.

### Architecture

The project consists of two main components:

- **WhatsApp Client**: Connects to WhatsApp using the whatsmeow library to fetch newsletter subscriber data
- **GitHub Updater**: Updates repository metadata and README files with current subscriber statistics

### Environment configuration

The service requires the following environment variables:

```env
# WhatsApp Newsletter Configuration
NUMBER_NEWSLETTERS=4
NEWSLETTER_JID1=120363394019833967@newsletter
NEWSLETTER_NAME1=Agrárias
NEWSLETTER_JID2=your_newsletter_jid
NEWSLETTER_NAME2=Central
NEWSLETTER_JID3=your_newsletter_jid
NEWSLETTER_NAME3=Botânico
NEWSLETTER_JID4=your_newsletter_jid
NEWSLETTER_NAME4=Politécnico

# GitHub Configuration
GITHUB_TOKEN=your_github_token
GITHUB_OWNER=guimox
GITHUB_REPO=ru-menu
```

### Technical stack

- **Go**: Primary programming language
- **whatsmeow**: WhatsApp Web API client library
- **SQLite**: Session storage for WhatsApp authentication
- **GitHub API**: Repository metadata updates
- **Docker**: Containerized deployment via GitHub Actions

### Project Structure

```
ru-counter/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── whatsapp/
│   │   └── client.go        # WhatsApp newsletter client
│   └── github/
│       └── updater.go       # GitHub repository updater
├── db/                      # SQLite session storage
├── docker-compose.yml       # Docker configuration
└── go.mod                   # Go dependencies
```

### Integration with RU Menu

This counter service is part of the larger [ru-menu](https://github.com/guimox/ru-menu) ecosystem that provides daily university restaurant menus to students via WhatsApp. The counter ensures accurate tracking of user engagement and service reach across multiple campus locations.

### Data Flow

1. **Authentication**: Service displays QR code for WhatsApp Web pairing
2. **Connection**: Establishes stable connection with WhatsApp servers
3. **Data Retrieval**: Fetches subscriber counts from configured newsletters
4. **GitHub Update**: Updates repository description and README with current statistics
5. **Reporting**: Logs successful updates with detailed breakdown

### Monitoring

The service provides detailed logging for:

- WhatsApp connection status and reconnection events
- Individual newsletter subscriber count retrieval
- GitHub API update operations
- Error handling and recovery procedures
