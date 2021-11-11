FROM golang:1.16.2-buster

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build

RUN cd /app/build/cmd/scheduler && go build -o /app/scheduler && cd /app/build && mv config-docker.json /app/config-docker.json && mv firebase.json /app/firebase.json
RUN rm -rf /app/build

ENTRYPOINT ["./scheduler", "--config=config-docker.json"]
