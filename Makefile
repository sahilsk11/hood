db-models:
	jet -dsn=postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable -path=./internal/db/models
	tools/env/bin/python3.9 tools/db_model_helper.py

db-drop-all:
	tools/env/bin/python3.9 tools/migrations.py down

db-run-all:
	tools/env/bin/python3.9 tools/migrations.py up

process-outfile:
	go run cmd/trade/main.go -command=process-outfile

update-prices:
	go run cmd/data-ingestion/main.go -command=update-prices

mocks:
	mockgen -source=internal/trade/trade_service.go -self_package=hood/internal/trade -package=trade > internal/trade/trade_service_mock.go

test-coverage:
	go test ./... -v -count=1 -coverprofile coverage.out > /dev/null
	gocover-cobertura < coverage.out > coverage.xml
	diff-cover coverage.xml --compare-branch=origin/master
	rm coverage.out
	rm coverage.xml