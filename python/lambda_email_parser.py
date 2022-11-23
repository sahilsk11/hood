import json
import re
from datetime import datetime
import boto3

def lambda_handler(event, context):
	trade = json.dumps(parse_event(event))
	pub_trade(trade)
	return {
		'statusCode': 200,
		'body': json.dumps('Hello from Lambda!')
	}

def parse_date(date):
	groups = re.search(r"([a-zA-Z]+) ([0-9]+).*, ([0-9]+) at (.*)", date).groups()
	cleaned_date = " ".join(list(groups))
	et_time_no_tz = datetime.strptime(cleaned_date, '%B %d %Y %I:%M %p')
	return et_time_no_tz

def parse_event(d):
	if len(d["Records"]) == 0:
		return
	message = json.loads(d["Records"][0]["Sns"]["Message"])
	content = message["content"]
	content = content.replace("\n", " ").replace("\r", " ").replace("  ", " ")
	groups = re.search(r"Your order to ([a-z]+) ([0-9]*[.][0-9]+) shares of ([A-Z]+) was executed at an average price of \$([0-9]*[.][0-9]+) on ([^\.]+).", content).groups()
	date = parse_date(groups[-1])
	return {
		"action": groups[0].upper(),
		"symbol": groups[2].upper(),
		"quantity": groups[1],
		"date": str(date),
		"note": "date is in ET timezone"
	}

def pub_trade(trade):
	client = boto3.client('sqs')
	client.send_message(
		QueueUrl="https://sqs.us-east-1.amazonaws.com/326651360928/hood-email-queue",
		MessageBody=(trade)
	)
