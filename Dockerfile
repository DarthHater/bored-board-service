# Copyright 2017 Jeffry Hesse

# Licensed under the Apache License, Version 2.0 (the "License"); 
# you may not use this file except in compliance with the License. 
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software 
# distributed under the License is distributed on an "AS IS" BASIS, 
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. 
# See the License for the specific language governing permissions and 
# limitations under the License. 

FROM golang:1.9

ARG app_env
ENV APP_ENV $app_env
ENV PRIVATE_KEY_PATH=/var/bored-board-service/.keys/app.rsa
ENV PUBLIC_KEY_PATH=/var/bored-board-service/.keys/app.rsa.pub

RUN apt-get update && apt-get install -y unzip openssl --no-install-recommends && \
    apt-get install -y supervisor && apt-get autoremove -y && apt-get clean -y && \
    wget -O dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 && \
    echo '322152b8b50b26e5e3a7f6ebaeb75d9c11a747e64bbfd0d8bb1f4d89a031c2b5 dep' | sha256sum -c - && \
    cp dep /usr/bin && rm dep

RUN chmod +x /usr/bin/dep

RUN mkdir -p /go/src/github.com/***
WORKDIR /go/src/github.com/***

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only -v

WORKDIR /go/src/github.com/DarthHater/bored-board-service

COPY . .

COPY .environment/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN chmod +x scripts/start.sh

RUN go build

CMD [ "/bin/bash", "scripts/start.sh" ]

EXPOSE 8000
EXPOSE 2345
