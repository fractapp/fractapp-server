FROM golang:1.16.2-buster

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build

RUN cd /app/build/cmd/api && go build -o /app/api && cd /app/build && mv config-docker.json /app/config-docker.json && mv assets /app/assets
RUN rm -rf /app/build

EXPOSE 9544
ENTRYPOINT ["./api", "--config=config-docker.json", "--host=0.0.0.0:9544"]
