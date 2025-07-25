# DNS Server with Web Management UI

A custom DNS server with a modern web interface for managing DNS records. Built with Go for the backend and React for the frontend.

## Features

### DNS Server
- 🌐 Custom DNS resolution with Redis storage
- 🔄 Automatic forwarding to upstream DNS (1.1.1.1) for unknown domains
- ⚡ High-performance UDP server with timeout handling

### Web Management Interface
- 📊 View all DNS records in a modern, responsive interface
- ➕ Add new DNS records with domain validation
- 🗑️ Delete existing DNS records
- 🔄 Real-time refresh functionality
- 📱 Mobile-friendly responsive design
- ⚠️ Error handling and user feedback

### REST API
- `GET /api/records` - List all DNS records
- `POST /api/records` - Create a new DNS record
- `DELETE /api/records/{domain}` - Delete a DNS record
- `GET /api/health` - Health check endpoint

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Go Backend    │    │     Redis       │
│   (React)       │◄──►│   DNS Server    │◄──►│   (Storage)     │
│   Port: 3000    │    │   Ports: 8080   │    │   Port: 6379    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Prerequisites

- Go 1.24+
- Bun (for frontend)
- Redis server
- Devbox (optional, for development environment)

## Setup and Installation

### 1. Start Redis Server
```bash
# using Docker (Recomemded)
docker run -d -p 6379:6379 redis:alpine

# But you can use your own redis setup but should be available only to localhost of the server
```

### 2. Backend Setup
```bash
# Install dependencies
go mod tidy

# Run the DNS server (requires sudo for port 53)
sudo go run cmd/dns/main.go
```

The server will start:
- DNS server on port 53 (UDP)
- REST API server on port 8080 (HTTP)

### 3. Frontend Setup
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
bun install

# Start development server
bun run dev
```

The web interface will be available at `http://localhost:3000`

## Usage

### Adding DNS Records via Web UI
1. Open `http://localhost:3000` in your browser
2. Click "Add DNS Record"
3. Enter domain name (e.g., `example.local`)
4. Enter IP address (e.g., `192.168.1.100`)
5. Optionally set TTL in seconds
6. Click "Add Record"

### Adding DNS Records via API
```bash
curl -X POST http://localhost:8080/api/records \
  -H "Content-Type: application/json" \
  -d '{"domain": "test.local", "ip": "192.168.1.50"}'
```

### Testing DNS Resolution
```bash
# Test with dig
dig @localhost test.local

# Test with nslookup
nslookup test.local localhost
```

### Viewing All Records
```bash
curl http://localhost:8080/api/records
```

### Deleting a Record
```bash
curl -X DELETE http://localhost:8080/api/records/test.local
```

## Configuration

### Redis Configuration
The Redis connection can be configured in `manager/redis.go`:
- Host: `localhost`
- Port: `6379`
- Database: `0`
- No authentication by default

### DNS Server Configuration
- Listening port: `53` (UDP)
- Upstream DNS: `1.1.1.1:53`
- Read timeout: `1 second`

### API Server Configuration
- Port: `8080`
- CORS enabled for all origins
- Request timeout: `10 seconds`

## Development

### Project Structure
```
├── main.go              # Main server entry point
├── api/                 # REST API handlers
│   └── api.go
├── handlers/            # DNS query handlers
│   └── handlers.go
├── manager/             # Redis connection manager
│   └── redis.go
├── constants/           # Global constants
│   └── constants.go
├── logger/              # Logging configuration
│   └── logger.go
├── utils/               # Utility functions
│   └── utils.go
└── frontend/            # React web interface
    ├── src/
    │   ├── App.jsx
    │   ├── App.css
    │   └── main.jsx
    └── package.json
```

### Running in Development
```bash
# Terminal 1: Start Redis
redis-server

# Terminal 2: Start Go backend
sudo go run main.go

# Terminal 3: Start React frontend
# note by default on starting backend at 3000 you can view the page
# if suppose if you want changes kindly consider to comment server.StartFrontend in main.go
# Before this step otherwise this step is not needed
cd frontend && bun run dev
```

## Production Deployment

### Frontend
```bash
cd frontend
bun vite build
cd -r ./dist ../
# Serve the dist/ directory with nginx or similar
```

### Backend
```bash
# Build the binary
go build -o dns cmd/dns/main.go

# Run with systemd or process manager
sudo ./dns-server
```



## Security Considerations

- The DNS server requires root privileges to bind to port 53
- Consider running with reduced privileges after binding
- Redis should be secured in production environments
- The API has no authentication - add auth for production use
- CORS is currently open to all origins

## Troubleshooting

### DNS Server Won't Start
- Check if port 53 is already in use: `sudo netstat -tulpn | grep :53`
- Ensure you're running with sudo privileges
- Verify Redis is running: `redis-cli ping`

### Web UI Can't Connect
- Verify API server is running on port 8080
- Check CORS settings if accessing from different domain
- Ensure Redis connection is established

### DNS Queries Not Working
- Test with `dig @localhost domain.local`
- Check DNS server logs for errors
- Verify records exist in Redis: `redis-cli keys "*"`

## License

MIT License - feel free to use and modify as needed.
