# Matchmaker

Matchmaker takes care of matching and planning of reviewers and review slots in people's calendars.

## Setup

You need to retrieve the `persons.yml` file containing people configuration for review.
Format example:
```yaml
- email: john.doe@kapten.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: chuck.norris@kapten.com
  isgoodreviewer: true
  maxsessionsperweek: 1
  skills:
    - frontend
    - data
- email: james.bond@kapten.com
- email: john.wick@kapten.com
  maxsessionsperweek: 1
- email: obi-wan.kenobi@kapten.com
```
**isgoodreviewer** [optional] is used to distinguish the experienced reviewers in order to create reviewer pairs that contain at least one experienced reviewer. Default value is false.
**maxsessionsperweek** [optional] sets a custom max sessions number per week for a reviewer. Default is 3. If set to 0, it also falls back to the default value.
**skills** [optional] describes the areas of expertise of a reviewer in order to create pairs of people with same competences. If not specified the reviewer can be paired with any other reviewer (no matter the skills)

You need to create/retrieve a `client_secret.json` file containing a valid Google Calendar
access token for your org's calendar.

Those files need to be placed at the root of the project.

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
