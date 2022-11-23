from raw import load
import json
import re
from datetime import datetime

def parse_date(date):
  print(date)
  groups = re.search(r"([a-zA-Z]+) ([0-9]+).*, ([0-9]+) at (.*)", date).groups()
  print(groups)
  # return datetime.strptime('Jun 1 2005  1:33PM', '%B %d %Y %I:%M%p')

def parse_context(d):
  if len(d["Records"]) == 0:
    return
  message = json.loads(d["Records"][0]["Sns"]["Message"])
  content = message["content"]
  content = content.replace("\n", " ").replace("\r", " ").replace("  ", " ")
  # print(content[content.index("Your order to"):content.index("Your order to")+300])
  groups = re.search(r"Your order to ([a-z]+) ([0-9]*[.][0-9]+) shares of ([A-Z]+) was executed at an average price of \$([0-9]*[.][0-9]+) on ([^\.]+).", content).groups()
  print(groups)
  parse_date(groups[-1])
  return

parse_context(load())