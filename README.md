# Matchmaker

Matchmaker takes care of matching and planning of reviewers and review slots in people's calendars.

## Install dependencies

Requires: [Go](https://golang.org/dl/) >= 1.22.x

    go install

## Build the app

    go build

## Install the app

    go install

## Run the app

You can now run the binary:

    ./matchmaker

# Updating dependencies

    go get -u
    go mod tidy

## Setup Google Calendar API Access

You need to setup a Google Cloud Platform project with the Google Calendar API enabled.
To create a project and enable an API, refer to [this documentation](https://developers.google.com/workspace/guides/create-project).

This simple app queries Google Calendar API as yourself, so you need to
have the authorization to create events and query availabilities on all the listed people's calendars.

You can follow the steps described [here](https://github.com/googleapis/google-api-nodejs-client#oauth2-client) to 
set up an OAuth2 client for the application.

Copy `configs/client_secret.json.example` into a new `configs/client_secret.json` file and replace values 
for `client_id`, `client_secret` and `project_id`.

## Get an access token

Once your credentials are set, you need to allow this app to use your
credentials. Just launch the command :

    matchmaker token

You should get a new browser window opening with a Google consent screen. If not you can 
open the url indicated in the command line :

    Authorize this app at: https://accounts.google.com/o/oauth2/auth?client_id=...

Grant access to the app by ignoring any security warning about the app not being verified.
Your token will be stored into a `calendar-api.json.json` file in your `~/.credentials` folder and a query to your
calendar will be made with it to test it, you should see the output in the console.

If you get an error :

    Response: {
      "error": "invalid_grant",
      "error_description": "Bad Request"
    }
    exit status 1


You need to delete the credential file `~/.credentials/calendar-api.json`:

    rm ~/.credentials/calendar-api.json

Then retry the command to create the token.

## Setup

You need to create/retrieve a group file `groups/group.yml` file containing people configuration for review.
Format example:
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
**isgoodreviewer** [optional] is used to distinguish the experienced reviewers in order to create reviewer pairs 
that contain at least one experienced reviewer. Default value is false.
**maxsessionsperweek** [optional] sets a custom max sessions number per week for a reviewer. Default is 3. 
If set to 0, it also falls back to the default value.
**skills** [optional] describes the areas of expertise of a reviewer in order to create pairs of people with 
same competences. If not specified the reviewer can be paired with any other reviewer (no matter the skills)

Copy the provided example file `group.yml.example` into a new `group.yml` file and replace values with actual users. 
You can have as many groups of people as you want, and name them as you want.

## Preparing

    matchmaker prepare [group-file [default=group.yml]] [--week-shift value [default=0]]

This command will compute work ranges for the target week, and check free slots for each potential
reviewer in the group file and create an output file `problem.yml`.

The group-file parameter specifies which group file to use from the groups directory.
You can create multiple group files (e.g., `teams.yml`, `projects.yml`) to manage different sets of people.
If no group file is specified, `group.yml` will be used by default.

By default, the command plans for the upcoming monday, you can provide a `weekShift` value as a parameter, allowing
to plan for further weeks (1 = the week after upcoming monday, etc.)

## Matching

    matchmaker match

This command will take input from the `problem.yml` file and match reviewers together in review slots for the target week.
The output is a `planning.yml` file with reviewers couples and planned slots.

## Planning

    matchmaker plan

This command will take input from the `planning.yml` file and create review events in reviewers' calendar.
