updateMocks:
	rm -r mocks/*
	${GOPATH}/bin/mockgen --source $(CURDIR)/db/db.go --destination $(CURDIR)/mocks/db/db_mock.go --package mocks
	${GOPATH}/bin/mockgen --source $(CURDIR)/notification/notificator.go --destination $(CURDIR)/mocks/notification/notificator_mock.go --package mocks
	${GOPATH}/bin/mockgen --source $(CURDIR)/push/notificator.go --destination $(CURDIR)/mocks/push/notificator_mock.go --package mocks
	@echo 'Mocks are updated'

totalCoverage:
	go test ./... -coverprofile=c.out
	go tool cover -func c.out | grep total

htmlCoverage:
	go test ./... -coverprofile=c.out && go tool cover -html=c.out

updateSwagger:
	${GOPATH}/bin/swag init -g cmd/api/main.go

build:
	mkdir -p bin
	rm -r bin
	mkdir -p bin
	cd bin && go build ../cmd/api && go build ../cmd/price && go build ../cmd/scheduler && go build ../cmd/subscriber
