FROM golang:1.16.2-buster

ENV currency="DOT"
ENV start=1597622400000

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build

RUN cd /app/build/cmd/price && go build -o /app/price && cd /app/build && mv config-docker.json /app/config-docker.json
RUN rm -rf /app/build

ENTRYPOINT ["./price", "--config=config-docker.json", "--currency=$currency", "--start=$start"]
