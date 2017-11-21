FROM golang:1.9

RUN apt-get update && apt-get install -y unzip --no-install-recommends && \
    apt-get autoremove -y && apt-get clean -y && \
    wget -O dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 && \
    echo '322152b8b50b26e5e3a7f6ebaeb75d9c11a747e64bbfd0d8bb1f4d89a031c2b5 dep' | sha256sum -c - && \
    cp dep /usr/bin && rm dep

RUN chmod +x /usr/bin/dep

WORKDIR /go/src/github.com/darthhater/bored-board-service
COPY . .

RUN dep ensure

EXPOSE 8000

CMD go run main.go
