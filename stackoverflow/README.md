# StackOverflow to Slack GitHub Action

This GitHub Action fetches new StackOverflow questions tagged with
`open-telemetry` and posts them to a Slack channel. It runs on an hourly
schedule.

## Setup

1. Clone this repository or create a new repository with the provided
   `action.py` script and `.github/workflows/get-questions.yml` workflow file.

2. Create a Slack [incoming webhook
   URL](https://api.slack.com/messaging/webhooks) for the Slack channel where
   you want to post new questions.

3. Add the webhook URL as a secret in your GitHub repository. Go to the
   "Settings" tab of your GitHub repository, then click "Secrets" in the left
   sidebar. Click "New repository secret" and add a secret with the name
   `SLACK_WEBHOOK_URL` and the webhook URL value.

4. Update the paths in the `.github/workflows/get-questions.yml` file to match
   the locations of your `action.py` file.

5. Push your changes to the repository. The GitHub Action will start running on an hourly schedule.

## How it works

The GitHub Action runs the `action.py` script, which fetches new StackOverflow
questions tagged with `open-telemetry` and posts them to a Slack channel.

In order to not repeat questions, the script stores the timestamp of the last
seen question in a file called `state.txt`. This file is cached by the GitHub
Actions runner and restored at the start of each run.

Since GitHub Actions does not allow you to update a cache key, the action relies
on the `restore-keys` behavior in `actions/cache`. This behavior checks for all
keys that match a prefix, then restores the _most recent_ one. Each run of the
action produces a new item (`state-<run id>`) in the cache, even if the
timestamp doesn't change.

In order to ensure that cache eviction occurs on our schedule,
`clean-cache.yaml` exists. Daily, it deletes all keys except for the most recent
one.

## Contributing / Possible Improvements

This should be pretty much done, but there's always improvements that could be
made. A fun project might be to generate metrics about the questions being asked
and publish them somewhere. Tags could be turned into links. There's a lot of
stuff that you get back in the search results that we aren't using.
[This page](https://api.stackexchange.com/docs/search#order=desc&sort=activity&tagged=open-telemetry&filter=default&site=stackoverflow&run=true)
shows you the response object, we're only using a few of the fields. It might be
nice to get the actual username, see how many answers exist already, etc.

I believe we could also use the Slack API better; there's a lot you can do with
it. See [this page](https://api.slack.com/reference/block-kit/blocks) for more
on that topic.
