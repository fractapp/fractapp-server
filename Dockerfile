FROM golang:rc

WORKDIR /app

RUN mkdir /app/build
COPY . /app/build

RUN cd /app/build && go build -o ../fractapp-server && mv config.release.json ../config.release.json && mv firebase.json ../firebase.json && mv assets ../assets
RUN rm -rf /app/build

EXPOSE 9544
ENTRYPOINT ["./fractapp-server", "--config=config.release.json", "--host=0.0.0.0:9544"]
