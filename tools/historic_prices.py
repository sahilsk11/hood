import csv
import psycopg2
from datetime import datetime, timezone

# Database connection details
db_config = {
    "host": "localhost",
    "database": "postgres",
    "user": "postgres",
    "password": "postgres",
    "port": "5438"
}

# CSV file path
csv_file = "result (1).csv"

# SQL statements
insert_query = "INSERT INTO price (symbol, price, updated_at, date) VALUES (%s, %s, %s, %s)"

# Open a connection to the database
conn = psycopg2.connect(**db_config)
cursor = conn.cursor()

# Read the CSV file and insert data into the database
with open(csv_file, "r", encoding="utf-8-sig") as file:
    reader = csv.DictReader(file)
    for row in reader:
        # print(row)
        symbol = row["TICKER"]  # Adjust the key name to include the BOM character
        if row["PRICECLOSE"] == "" and symbol != "SNOXX":
            print(row)
            continue
        elif row["PRICECLOSE"] == "":
          row["PRICECLOSE"] = 1
        else:
          price = float(row["PRICECLOSE"]) 
        date = row["PRICEDATE"]

        if symbol == "SNOXX":
            price = 1.0

        updated_at = datetime.now(timezone.utc).isoformat()

        cursor.execute(insert_query, (symbol, price, updated_at, date))

# Commit the changes and close the connection
conn.commit()
cursor.close()
conn.close()
