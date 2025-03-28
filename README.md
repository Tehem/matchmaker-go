# Matchmaker

Matchmaker takes care of matching and planning of reviewers and review slots in people's calendars.

## Project Structure

```
matchmaker/
├── cmd/
│   └── matchmaker/     # Main application entry point
├── internal/
│   ├── calendar/       # Google Calendar API integration
│   ├── commands/       # CLI commands
│   ├── config/        # Configuration management
│   └── matching/      # Reviewer matching logic
├── configs/           # Configuration files
└── pkg/              # Public packages (if any)
```

## Requirements

- [Go](https://golang.org/dl/) >= 1.22.x

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/matchmaker-go.git
   cd matchmaker-go
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o matchmaker ./cmd/matchmaker
   ```

4. Install the application:
   ```bash
   go install ./cmd/matchmaker
   ```

## Configuration

### Application Configuration

The application uses a YAML configuration file located at `configs/config.yml`. You can create it from the example:

```bash
cp configs/config.yml.example configs/config.yml
```

The configuration includes:
- Session duration and spacing
- Maximum sessions per person per week
- Working hours and timezone
- Session prefix for calendar events

### People Configuration

Create a `persons.yml` file with the list of reviewers. Example:
```yaml
- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: chuck.norris@example.com
  isgoodreviewer: true
  maxsessionsperweek: 1
  skills:
    - frontend
    - data
- email: james.bond@example.com
- email: john.wick@example.com
  maxsessionsperweek: 1
- email: obi-wan.kenobi@example.com
```

Configuration options:
- `isgoodreviewer` [optional]: Marks experienced reviewers for better pairing
- `maxsessionsperweek` [optional]: Sets custom max sessions per week (default: 3)
- `skills` [optional]: Areas of expertise for skill-based matching

## Google Calendar Setup

1. Create a Google Cloud Platform project and enable the Google Calendar API
2. Create OAuth 2.0 credentials and download them as `client_secret.json`
3. Get an access token:
   ```bash
   matchmaker token
   ```
   This will open a browser window for authorization and save the token in `~/.credentials/calendar-api.json`

## Usage

### Prepare

Compute work ranges and check free slots for the target week:
```bash
matchmaker prepare [--week-shift value]
```
- `--week-shift`: Number of weeks to shift from current week (default: 0)
- Output: `problem.yml`

### Match

Match reviewers together in review slots:
```bash
matchmaker match
```
- Input: `problem.yml`
- Output: `planning.yml`

### Plan

Create review events in reviewers' calendars:
```bash
matchmaker plan
```
- Input: `planning.yml`
- Creates calendar events for all matched reviewers

## Development

### Updating Dependencies

```bash
go get -u
go mod tidy
```

### Running Tests

```bash
go test ./...
```

### Code Style

The project follows standard Go formatting. Format your code before committing:
```bash
go fmt ./...
```

## License

[Your chosen license]
