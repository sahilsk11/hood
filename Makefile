db-models:
	jet -dsn=postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable -path=./internal/db/models
	python3 /Users/skapur/portfolio/hood/tools/db_model_helper.py

db-drop-all:
	tools/env/bin/python3.9 tools/migrations.py down

db-run-all:
	tools/env/bin/python3.9 tools/migrations.py up

process-outfile:
	go run cmd/data-ingestion/main.go -command=process-outfile

update-prices:
	go run cmd/data-ingestion/main.go -command=update-prices