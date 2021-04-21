
updateMocks:
	rm mocks/*
	${GOPATH}/bin/mockgen --source $(CURDIR)/db/db.go --destination $(CURDIR)/mocks/db_mock.go --package mocks
	${GOPATH}/bin/mockgen --source $(CURDIR)/notification/notificator.go --destination $(CURDIR)/mocks/notificator_mock.go --package mocks
	${GOPATH}/bin/mockgen --source $(CURDIR)/firebase/notificator.go --destination $(CURDIR)/mocks/firebase_mock.go --package mocks
	${GOPATH}/bin/mockgen --source $(CURDIR)/adaptors/adaptor.go --destination $(CURDIR)/mocks/adaptor_mock.go --package mocks
	@echo 'Mocks are updated'

totalCoverage:
	go test ./... -coverprofile=c.out
	go tool cover -func c.out | grep total

htmlCoverage:
	go test ./... -coverprofile=c.out && go tool cover -html=c.out

updateSwagger:
	${GOPATH}/bin/swag init -g cmd/api/main.go

