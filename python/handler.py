"""
The existing code is terrible to read
and doesn't really make sense. I'm keeping it
until I fully understand and can break it down
further. In the meantime, here are some
high level functions to help process data
"""

from rh import parse_row, DecimalEncoder
import json

inp = """
SPDR S&P 500 ETF Sell
Mar 6
$5,000.00
12.279503 shares at $407.18
Symbol
SPY
Type
Sell
Time in Force
Good for day
Submitted
Mar 6, 2023
Status
Filled
Entered Amount
$5,000.00
Filled
Mar 6, 2023 at 12:20 PM EST
Filled Quantity
12.279503 shares at $407.18
Filled Notional
$5,000.00

NVIDIA Market Sell
May 31
$4,999.41
13.114754 shares at $381.21
Symbol
NVDA
Type
Market Sell
Time in Force
Good for day
Submitted
May 31, 2023
Status
Filled
Entered Amount
$5,000.00
Filled
May 31, 2023 at 1:01 PM EDT
Filled Quantity
13.114754 shares at $381.21
Filled Notional
$4,999.41
"""

rows = inp.split("\n\n")
rows = [x.strip().split("\n") for x in rows]
print("this", rows)

for r in rows:
  print(json.dumps(parse_row(r), cls=DecimalEncoder))