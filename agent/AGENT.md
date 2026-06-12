# project goal
Create single-page web app where users can find and upload alternative images of existing magic the gathering cards. Read agent/summary.md to get up to speed with the current status of the code

# bootstrapping
- read @agent/summary.md for code overview
- never build manually, assume `air` is running, if it is not, run `air`
- server url: http://localhost:8080/frontend
- say "done" once you have finished all tasks and all necessary checks succeeded

# values
## simplicity - less is better
- less files
- less dependencies
- less tools
- less code: BUT code must be easily reabable for humans, use concise & descriptive names for variables/functions, code should be self explanatory - if not possible (only then!) use comments
- minimize number of steps/clicks for users to achieve their goal

## explicitness
- use abbreviations sparingly (example: "single page app" good, "SPA" bad)
- exception: unabbreviated form is unusual/harder to understand/clunky (example: "app" good, "application" bad; "laser" good, "light amplification by stimulated emission of radiation" bad)

## uncertainty
- if uncertain about your answer to direct question, give answer, but emphasize uncertainty; if certain do not comment on (un)certainty

# constraints

## clarity
- before getting to work, check if anything is unclear from the prompt
- if yes, create a list of open questions, for each find 2 possible answers (without explanation)
- decide if the answers to these questions actually matter
- if no, just pick whatever answer is better
- if yes, instead of getting to work, report the list of questions and possible answers

## test first
before implementing any new logic or a new route in a web app:
1. write test, run tests (`go test ./server/tests/`), new test must fail; don't continue unless tests ran and new test failed
2. implement, run tests, all must succeed; don't continue/finish unless all tests succeed, including test from step 1

## clean code
when writing a file don't finish until there's
- no trailing whitespace
- newline at end of file

## update agent docs
- dont finish until @agent/summary.md was updated to reflect new project state
