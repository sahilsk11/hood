import os

os.chdir("/Users/skapur/portfolio/hood/internal/db/models/postgres/public/model/")
files = os.listdir()
for file in files:
  f = open(file)
  contents = f.read()
  f.close()
  if "float64" in contents:
    contents = contents.replace("float64", "decimal.Decimal")
    if "\"time\"" in contents:
      contents = contents.replace("\"time\"", "\"time\"\n\n\t\"github.com/shopspring/decimal\"")
    else:
      contents = contents.replace("package model", "package model\n\nimport (\n\t\"github.com/shopspring/decimal\"\n)")
    f = open(file, 'w')
    f.write(contents)
    f.close()
