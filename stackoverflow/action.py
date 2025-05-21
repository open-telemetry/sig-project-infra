import os
import json
import requests
import logging
from datetime import datetime

STACK_OVERFLOW_API = "https://api.stackexchange.com/2.3/search"
SLACK_WEBHOOK_URL = os.environ.get('SLACK_WEBHOOK_URL')
STATE_FILE = "state.txt"

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def get_latest_question_timestamp():
    try:
        with open(STATE_FILE, 'r') as f:
            content = f.read().strip()
            return int(content) if content else 0
    except FileNotFoundError:
        return 0

def update_latest_question_timestamp(latest_timestamp):
    with open(STATE_FILE, 'w') as f:
        f.write(str(latest_timestamp))

def fetch_questions(latest_timestamp):
    params = {
        "tagged": "open-telemetry;otel",
        "sort": "creation",
        "site": "stackoverflow"
    }
    if latest_timestamp > 0:
        params["min"] = latest_timestamp + 1

    response = requests.get(STACK_OVERFLOW_API, params=params)
    response.raise_for_status()
    return response.json().get('items', [])

def format_question(question):
    title = question['title']
    link = question['link']
    tags = [tag for tag in question['tags'] if tag != 'open-telemetry']
    tags_str = ', '.join(tags)
    message = {
        "blocks": [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": "New StackOverflow Question",
                    "emoji": True
                }
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"<{link}|{title}>\n*Tags:* {tags_str}"
                }
            },
            {
                "type": "divider"
            },
            {
                "type": "context",
                "elements": [
                    {
                        "type": "mrkdwn",
                        "text": f"Posted at {datetime.fromtimestamp(question['creation_date']).strftime('%Y-%m-%d %H:%M:%S')}"
                    }
                ]
            }
        ],
        "text": title
    }
    return message

def post_to_slack(message):
    if SLACK_WEBHOOK_URL:
        headers = {"Content-Type": "application/json"}
        data = {"text": message['text'], "blocks": message['blocks']}
        response = requests.post(SLACK_WEBHOOK_URL, headers=headers, data=json.dumps(data))
        response.raise_for_status()
    else:
        logging.info("SLACK_WEBHOOK_URL is not set. Logging the results instead.")
        logging.info(f"Message: {message}")

def main():
    latest_timestamp = get_latest_question_timestamp()
    logging.info(f"Latest question timestamp: {latest_timestamp}")
    questions = fetch_questions(latest_timestamp)
    logging.info(f"Fetched {len(questions)} questions from Stack Overflow")

    for question in questions:
        message = format_question(question)
        logging.info(f"Formatting question: {question['title']}")
        post_to_slack(message)
        logging.info(f"Posted question to Slack: {question['title']}")
        latest_timestamp = max(latest_timestamp, question['creation_date'])

    update_latest_question_timestamp(latest_timestamp)
    logging.info(f"Updated latest question timestamp: {latest_timestamp}")

if __name__ == "__main__":
    main()
