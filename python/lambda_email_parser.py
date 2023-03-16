import json
import re
from datetime import datetime

def lambda_handler(event, context):
	try:
		parse_and_save_event(event)
	except Exception as e:
		print(event)
		raise e
	

def parse_date(date):
	groups = re.search(r"([a-zA-Z]+) ([0-9]+).*, ([0-9]+) at (.*)", date).groups()
	cleaned_date = " ".join(list(groups))
	et_time_no_tz = datetime.strptime(cleaned_date, '%B %d %Y %I:%M %p')
	return et_time_no_tz

# brings tda message from 34k chars to 8k
def clean_email_content(content):
	content = content.replace("\n", " ").replace("\r", " ").replace("  ", " ")
	content = re.sub(r"style=3D\".*\"","", content, flags=re.IGNORECASE)
	return content


def lambda_event_to_email(d):
	if len(d["Records"]) == 0:
		raise Exception("no records found")
	message = json.loads(d["Records"][0]["Sns"]["Message"])
	if "mail" not in message.keys():
		raise Exception("invalid message type")
	source = message["mail"]["source"].lower()
	destination = message["mail"]["destination"][0].lower()
	subject = message["mail"]["commonHeaders"]["subject"]
	content =clean_email_content(message["content"])
	return {
		"sender": source,
		"subject": subject,
		"content": content,
		"receiver": destination
	}

def validate_email(email):
	if email["sender"] != "sahilkapur.a@gmail.com":
		raise Exception("unrecognized email sender "+email["sender"])
	return None
	
def parse_rh_email(email):
	groups = re.search(r"Your order to ([a-z]+) ([0-9]*[.][0-9]+) shares of ([A-Z]+) was executed at an average price of \$([0-9]*[.][0-9]+) on ([^\.]+).", email["content"]).groups()
	date = parse_date(groups[4])
	return {
		"action": groups[0].upper(),
		"symbol": groups[2].upper(),
		"quantity": groups[1],
		"cost_basis": groups[3],
		"date": str(date),
		"note": "date is in ET timezone"
	}

def parse_and_save_event(d):
	email = lambda_event_to_email(d)
	validate_email(email)
	print(email)


def pub_trade(trade):
	client = boto3.client('sqs')
	client.send_message(
		QueueUrl="https://sqs.us-east-1.amazonaws.com/326651360928/hood-email-queue",
		MessageBody=(trade)
	)
