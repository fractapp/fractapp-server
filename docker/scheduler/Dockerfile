FROM golang:1.16.2-buster

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build
COPY config-docker.json /app/config-docker.json
COPY firebase.json /app/firebase.json

RUN cd /app/build/cmd/scheduler && go build -o /app/scheduler && cd /app/build
RUN rm -rf /app/build

ENTRYPOINT ["./scheduler", "--config=config-docker.json"]
