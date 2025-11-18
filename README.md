# RU Counter

Automated WhatsApp newsletter subscriber counter for the [ru-menu](https://github.com/guimox/ru-menu) project. This service automatically fetches subscriber counts from WhatsApp newsletters and updates GitHub repository statistics.

### Overview

RU Counter is a microservice that connects to WhatsApp Business API to retrieve real-time subscriber counts from university restaurant menu newsletters. It automatically updates the main repository description and README with current Daily Active User (DAU) statistics.

### Architecture

The project consists of two main components:

- **WhatsApp Client**: Connects to WhatsApp using the whatsmeow library to fetch newsletter subscriber data
- **GitHub Updater**: Updates repository metadata and README files with current subscriber statistics

### Project Structure

```
ru-counter/
├── cmd
│   └── main.go
├── db
│   └── session.db
├── docker-compose-prd.yml
├── Dockerfile
├── go.mod
├── go.sum
├── internal
│   ├── github
│   │   └── updater.go
│   └── whatsapp
│       └── client.go
└── README.md
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
