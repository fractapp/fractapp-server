FROM golang:1.16.2-buster

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build

RUN cd /app/build/cmd/subscriber && go build -o /app/subscriber && cd /app/build && mv config-docker.json /app/config-docker.json
RUN rm -rf /app/build

EXPOSE 9505
ENTRYPOINT ["./subscriber", "--config=config-docker.json", "--host=0.0.0.0:9505"]
