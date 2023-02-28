from holdings import holdings as raw_trades
import calendar
import json
from decimal import *
import datetime

class DecimalEncoder(json.JSONEncoder):
  def default(self, obj):
    if isinstance(obj, Decimal):
        return float(obj)
    return json.JSONEncoder.default(self, obj)

rows = raw_trades.split("\n\n")
rows = [x.strip().split("\n") for x in rows]

trades = []
holdings = {
  "cash": Decimal(0)
}

def format_date(raw_date):
  raw_date = raw_date.replace(",", "")
  data = raw_date.split(" ")
  year = "2022"
  if len(data) > 2:
    year = data[2]
  day = data[1]
  month = 1
  while calendar.month_abbr[month] != data[0]:
    month += 1
  month = str(month)
  if len(month) == 1:
    month = "0"+month
  if len(day) == 1:
    day = "0"+day
  return "-".join([year, str(month), day])

def parse_purchase_details(details):
  details = details.replace(",", "")
  details = details.split(" ")
  num_shares = Decimal(details[0])
  price = Decimal(details[-1].replace("$", ""))
  return (num_shares, price)

# abs(d2 - d1)
def days_between(d1, d2):
  d2_attr = d2.split("-")
  d1_attr = d1.split("-")
  b = datetime.date(int(d2_attr[0]), int(d2_attr[1]), int(d2_attr[1]))
  a = datetime.date(int(d1_attr[0]), int(d1_attr[1]), int(d1_attr[1]))
  return abs((b-a).days)

def verify_row(row):
  x = None
  if row[1] not in row[17]:
    if input("mismatched dates " + str(row)) != "c":
      exit(1)
  if row[3] != row[19]:
    if input("mismatched symbols " + str(row)) != "c":
      exit(1)
  return

def parse_row(row):
  if len(row) < 3:
    return
  if row[2] == "Canceled" or row[2] == "Failed" or row[2] == "Rejected":
    return
  header = row[0].upper()
  if len(row) > 10 and ("LIMIT" in header or "STOP LOSS" in header):
    row.pop(10)
    row.pop(10)
  header = header.replace(" STOP LOSS", "").replace(" TRAILING STOP", "").replace(" LIMIT", "")
  date = format_date(row[1])
  long_date = None
  if len(row) > 17:
    long_date = row[17]
  if len(row) > 3:
    if "at $" in row[3]:
      (num_shares, cost_basis) = parse_purchase_details(row[3])
    if row[3] == "Reinvested":
      return
  

  out = {
    "date": date
  }

  if "FREE CHECKING" in header:
    # out["asset"] = "CASH"
    # out["quantity"] = nominal_price
    # out["cost_basis"] = Decimal(1)
    # if "DEPOSIT" in header:
    #   out["description"] = "BANK_DEPOSIT"
    # if "WITHDRAWAL" in header:
    #   out["description"] = "BANK_WITHDRAWAL"
    #   out["cost_basis"] *= Decimal(-1)
    # return out
    return

  if "FORWARD SPLIT" in header:
    out["asset"] = header.replace(" FORWARD SPLIT", "")
    out["action"] = "SPLIT"
    split = row[2].split(":")
    out["ratio"] = int(split[0])
    return out

  if "BONUS STOCK" in header:
    out["asset"] = "SIRIUS XM"
    out["action"] = "BUY"
    out["quantity"] = Decimal(1)
    out["cost_basis"] = 0
    return out

  if "BUY" in header:
    # verify_row(row)
    asset_name = header.replace(" BUY", "")
    out["asset"] = asset_name
    if asset_name == "ETHEREUM" or asset_name == "BITCOIN" or asset_name == "DOGECOIN":
      long_date = row[13].replace("2022,", "2022").replace("2021,", "2021").replace("2020,", "2020")
    out["action"] = "BUY"
    out["quantity"] = num_shares
    out["cost_basis"] = cost_basis
    out["description"] = "BUY "+asset_name
    if long_date == None:
      print("long date missing. you need to expand the stock details then copy it over")
      print(out)
      exit(1)
    out["long_date"] = long_date.replace(" at", "")
    return out
  if "SELL" in header:
    # verify_row(row)
    asset_name = header.replace(" SELL", "")
    if asset_name == "ETHEREUM" or asset_name == "BITCOIN" or asset_name == "DOGECOIN":
      long_date = row[13].replace("2022,", "2022").replace("2021,", "2021").replace("2020,", "2020")
    out["asset"] = asset_name
    out["action"] = "SELL"
    out["quantity"] = num_shares
    out["cost_basis"] = cost_basis
    out["description"] = "SELL "+asset_name
    if long_date == None:
      print("long date missing")
      print(out)
      exit(1)
    out["long_date"] = long_date.replace(" at", "")
    return out


  if "DIVIDEND FROM" in header:
    # out["asset"] = "CASH"
    # out["quantity"] = nominal_price
    # out["cost_basis"] = Decimal(1)
    # out["description"] = header
    # return out
    return

  if "INTEREST" in header:
    # out["asset"] = "CASH"
    # out["quantity"] = nominal_price
    # out["cost_basis"] = Decimal(1)
    # out["description"] = header
    # return out
    return

  if "ROBINHOOD GOLD" in header:
    # out["asset"] = "CASH"
    # out["quantity"] = nominal_price
    # out["cost_basis"] = Decimal(-1)
    # out["description"] = header
    # return out
    return

  print("unhandled case", header)
  exit(1)

for row in rows:
  parsed_row = parse_row(row)
  if parsed_row != None:
    trades.append(parsed_row)

print(json.dumps(trades, cls=DecimalEncoder))
exit(0)

assets = {

}

def count_cash(trades):
  cash = Decimal(0)
  # need to maintain original ordering, but reversed
  # tbh should be no need to sort after reversing
  date_sorted_trades = reversed(trades)
  for trade in date_sorted_trades:
    if "action" in trade and trade["action"] == "BUY":
      cash -= trade["quantity"]*trade["cost_basis"]
    elif "action" in trade and trade["action"] == "SELL":
      cash += trade["quantity"]*trade["cost_basis"]
    elif "asset" in trade and trade["asset"] == "CASH":
      # negative cost basis is for RH gold
      cash += trade["quantity"] * trade["cost_basis"]


  print(cash)

def count_holdings(trades):
  date_sorted_trades = reversed(trades)
  for trade in date_sorted_trades:
    asset = trade["asset"]
    if asset not in assets:
      assets[asset] = 0
    if "action" in trade and trade["action"] == "BUY":
       assets[asset] += trade["quantity"]
    elif "action" in trade and trade["action"] == "SELL":
      assets[asset] -= trade["quantity"]
    elif "action" in trade and trade["action"] == "SPLIT":
      print(asset, assets[asset])
      assets[asset] *= trade["ratio"]
      print(assets[asset])
    
    if assets[asset] < 0.00000000001:
      del assets[asset]

  print(json.dumps(assets))

def construct_lots(date_sorted_trades):
  lots = {}
  for trade in date_sorted_trades:
    asset = trade["asset"]

    if "action" in trade and trade["action"] == "BUY":
      if asset not in lots:
        lots[asset] = {
          "total_quantity": Decimal(0),
          "open_lots": []
        }
      lots[asset]["open_lots"].append({
        "quantity": trade["quantity"],
        "date": trade["date"],
        "cost_basis": trade["cost_basis"]
       })
      lots[asset]["total_quantity"] += trade["quantity"]
    elif "action" in trade and trade["action"] == "SELL":
      amount_to_sell = trade["quantity"]
      while amount_to_sell > 0:
        lot = lots[asset]["open_lots"][0]
        sold_from_lot = min(lot["quantity"], amount_to_sell)
        gains = (trade["cost_basis"]-lot["cost_basis"])*sold_from_lot
        
        if trade["date"][:4] == "2022":
          hold_duration = days_between(trade["date"], lot["date"])
          if hold_duration >= 365:
            if "realized_lt_gains" not in lots[asset]:
              lots[asset]["realized_lt_gains"] = 0
            lots[asset]["realized_lt_gains"] += gains
          else:
            if "realized_st_gains" not in lots[asset]:
              lots[asset]["realized_st_gains"] = 0
            lots[asset]["realized_st_gains"] += gains

        amount_to_sell -= sold_from_lot
        lots[asset]["total_quantity"] -= sold_from_lot
        lot["quantity"] -= sold_from_lot
        if lot["quantity"] == 0:
          lots[asset]["open_lots"].pop(0)
        if len(lots[asset]["open_lots"]) == 0:
          del lots[asset]

    elif "action" in trade and trade["action"] == "SPLIT":
      lots[asset]["total_quantity"] *= trade["ratio"]
      asset_lots = lots[asset]["open_lots"]
      for l in asset_lots:
        l["cost_basis"] /= trade["ratio"]
        l["quantity"] *= trade["ratio"]
  return lots

def add_gains(lots):
  for symbol in lots.keys():
    unrealized_st_gains = 0
    unrealized_lt_gains = 0
    today = "2022-11-17"
    for lot in lots[symbol]["open_lots"]:
      current_price = get_price(symbol)
      gains = (current_price-lot["cost_basis"])*lot["quantity"]
      hold_duration = days_between(today, lot["date"])
      if hold_duration >= 365:
        unrealized_lt_gains += gains
      else:
        unrealized_st_gains += gains
      lots[symbol]["unrealized_st_gains"] = unrealized_st_gains
      lots[symbol]["unrealized_lt_gains"] = unrealized_lt_gains
  return lots

def calculate_breakeven_price_change(gains, quantity, current_price):
  tlh_gain = -gains*Decimal(0.15)
  # to save the same amount by holding, asset
  # needs to go up by the same amount AFTER
  # deducting taxes
  equivalent_gain = tlh_gain/Decimal(0.85)
  breakeven_gain_per_share = equivalent_gain / quantity
  return (breakeven_gain_per_share / current_price) * Decimal(100)

def add_breakeven_percentages(lots):
  for symbol in lots.keys():
    quantity = lots[symbol]["total_quantity"]
    current_price = get_price(symbol)
    gains = lots[symbol]["unrealized_st_gains"] + lots[symbol]["unrealized_lt_gains"]
    lots[symbol]["risk_percentage"] = calculate_breakeven_price_change(gains, quantity, current_price)
  return lots

def idk():
  date_sorted_trades = reversed(trades)
  lots = construct_lots(date_sorted_trades)
  lots = add_gains(lots)
  lots = add_breakeven_percentages(lots)
  # for symbol in lots:

  #   del lots[symbol]["open_lots"]
  print(json.dumps(lots, cls=DecimalEncoder))

# print([x for x in trades if "asset" in x and x["asset"] == "CASH" and x["cost_basis"] != Decimal(1)])
# idk()
# count_cash(trades)
# count_holdings(trades)
# print(json.dumps(trades))
