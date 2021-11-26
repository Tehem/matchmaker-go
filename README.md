# Matchmaker

Matchmaker takes care of matching and planning of reviewers and review slots in people's calendars.

## Install dependencies

Requires: [Go](https://golang.org/dl/) >= 1.17.x

    go install


## Setup Google Calendar API Access

You need to setup a Google Cloud Platform project with the Google Calendar API enabled.
To create a project and enable an API, refer to [this documentation](https://developers.google.com/workspace/guides/create-project).

This simple app queries Google Calendar API as yourself, so you need to
have the authorization to create events and query availabilities on all
of the listed people's calendars.

You can follow the steps described [here](https://github.com/googleapis/google-api-nodejs-client#oauth2-client) to 
setup an OAuth2 clien for the application.

Copy `client_secret.json.example` into a new `client_secret.json` file and
replace values for `client_id`, `client_secret` and `project_id` (id
de client OAuth).

## Get an access token

Once your credentials are set, you need to allow this app to use your
credentials. Just launch the script :

    go run quickstart.go


If you get an error :

    Response: {
      "error": "invalid_grant",
      "error_description": "Bad Request"
    }
    exit status 1


You need to delete the credential file `~/.credentials/calendar-api.json`:

    rm ~/.credentials/calendar-api.json

Then retry the script to create the token.

You should get a prompt to open an URL like this:

    Go to the following link in your browser then type the authorization code:: https://accounts.google.com/o/oauth2/v2/auth?access_type=offline&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcalendar.readonly&response_type=code&client_id=xxx.apps.googleusercontent.com&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob


Grant access and paste the provided code in the command line and press enter.
Your token will be stored into a `calendar-api.json.json` file in your `~/.credentials` folder and a query to your
calendar will be made with it to test it, you should see the output.

## Setup

You need to create/retrieve the `persons.yml` file containing people configuration for review.
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
**isgoodreviewer** [optional] is used to distinguish the experienced reviewers in order to create reviewer pairs that contain at least one experienced reviewer. Default value is false.
**maxsessionsperweek** [optional] sets a custom max sessions number per week for a reviewer. Default is 3. If set to 0, it also falls back to the default value.
**skills** [optional] describes the areas of expertise of a reviewer in order to create pairs of people with same competences. If not specified the reviewer can be paired with any other reviewer (no matter the skills)

Copy the provided example file `persons.yml.example` into a new `persons.yml` file and replace values with actual users.

TEMP / TO REMOVE :
Adapt the master email in this file : `plan.go:34`


## Preparing

    go run prepare.go [-week-shift value [default=0]]

This script will compute work ranges for the target week, and check free slots for each potential
reviewer and create an output file `problem.yml`.

By default, the script plans for the upcoming monday, you can provide a `weekShift` value as a parameter, allowing
to plan for further weeks (1 = the week after upcoming monday, etc.)

## Matching

    go run match.go

This script will take input from the `match.yml` file and match reviewers together in review slots for the target week.
The output is a `planning.yml` file with reviewers couples and planned slots.

## Planning

    go run plan.go

This script will take input from the `planning.yml` file and create review events in reviewers' calendar.


## Default run

By running the script:

    ./do.sh

All scripts will be run sequentially for the upcoming monday.
