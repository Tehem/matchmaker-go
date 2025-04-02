# 🎯 Matchmaker

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Matchmaker takes care of matching coworkers and planning review or pairing slots in people's calendars.

## 📋 Table of Contents
- [Installation](#-installation)
- [Setup](#-setup)
- [Commands](#-commands)
- [Google Calendar API Setup](#-google-calendar-api-setup)
- [Usage Examples](#-usage-examples)

## 🚀 Installation

### Prerequisites
- [Go](https://golang.org/dl/) >= 1.22.x

### Install Dependencies
```bash
go install
```

### Build the App
```bash
go build
```

### Install the App
```bash
go install
```

### Run the App
```bash
./matchmaker
```

### Update Dependencies
```bash
go get -u
go mod tidy
```

## ⚙️ Setup

You need to create/retrieve a group file `groups/group.yml` containing people configuration for review.

### Group File Format Example
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

### Configuration Options
- **isgoodreviewer** [optional] - Used to distinguish experienced reviewers to create pairs with at least one experienced reviewer. Default: `false`
- **maxsessionsperweek** [optional] - Sets a custom max sessions number per week for a reviewer. Default: `3`. If set to `0`, it falls back to the default value.
- **skills** [optional] - Describes areas of expertise to create pairs with same competences. If not specified, the reviewer can be paired with any other reviewer.

Copy the provided example file `group.yml.example` into a new `group.yml` file and replace values with actual users. You can have as many groups of people as you want, and name them as you want.

## 🛠️ Commands

### 🔍 Prepare
```bash
matchmaker prepare [group-file [default=group.yml]] [--week-shift value [default=0]]
```

This command computes work ranges for the target week, checks free slots for each potential reviewer in the group file, and creates an output file `problem.yml`.

- **group-file**: Specifies which group file to use from the groups directory
- **--week-shift**: Plans for further weeks (1 = the week after upcoming Monday, etc.)

### 🤝 Match
```bash
matchmaker match
```

This command takes input from the `problem.yml` file and matches reviewers together in review slots for the target week. The output is a `planning.yml` file with reviewer couples and planned slots.

### 📅 Plan
```bash
matchmaker plan [file]
```

This command takes input from a planning file and creates review events in reviewers' calendars.

- If no file is specified, it will:
  - Use `planning.yml` if it's the only file present
  - Use `weekly-planning.yml` if it's the only file present
  - Ask which file to use if both are present
- You can also specify a file directly: `matchmaker plan my-planning.yml`

### 🔄 Weekly Match
```bash
matchmaker weekly-match [group-file]
```

This command creates random pairs of people with no common skills and schedules sessions across consecutive weeks.

- Takes a group file as input (default: `group.yml`)
- Ensures paired people have no common skills
- Schedules sessions with optimal timing preferences
- Outputs a `weekly-planning.yml` file with all scheduled sessions

## 🔐 Google Calendar API Setup

You need to setup a Google Cloud Platform project with the Google Calendar API enabled.
To create a project and enable an API, refer to [this documentation](https://developers.google.com/workspace/guides/create-project).

This simple app queries Google Calendar API as yourself, so you need to
have the authorization to create events and query availabilities on all the listed people's calendars.

You can follow the steps described [here](https://github.com/googleapis/google-api-nodejs-client#oauth2-client) to 
set up an OAuth2 client for the application.

Copy `configs/client_secret.json.example` into a new `configs/client_secret.json` file and replace values 
for `client_id`, `client_secret` and `project_id`.

## 🔑 Get an Access Token

Once your credentials are set, you need to allow this app to use your
credentials. Just launch the command:

```bash
matchmaker token
```

You should get a new browser window opening with a Google consent screen. If not you can 
open the url indicated in the command line:

```
Authorize this app at: https://accounts.google.com/o/oauth2/auth?client_id=...
```

Grant access to the app by ignoring any security warning about the app not being verified.
Your token will be stored into a `calendar-api.json.json` file in your `~/.credentials` folder and a query to your
calendar will be made with it to test it, you should see the output in the console.

If you get an error:

```
Response: {
  "error": "invalid_grant",
  "error_description": "Bad Request"
}
exit status 1
```

You need to delete the credential file `~/.credentials/calendar-api.json`:

```bash
rm ~/.credentials/calendar-api.json
```

Then retry the command to create the token.

## 📝 Usage Examples

### Basic Workflow
```bash
# Prepare for next week
matchmaker prepare

# Match reviewers
matchmaker match

# Create calendar events
matchmaker plan
```

### Weekly Match Workflow
```bash
# Create random pairs and schedule sessions
matchmaker weekly-match groups/my-team.yml

# Create events from the generated planning file
matchmaker plan
```

### Advanced Usage
```bash
# Prepare for a specific week (2 weeks from now)
matchmaker prepare --week-shift 2

# Use a specific group file
matchmaker prepare groups/backend-team.yml

# Specify a planning file directly
matchmaker plan my-custom-planning.yml
```
