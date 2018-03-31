#!/bin/sh

# Build on changes to source unless production
if [[ $APP_ENV = production ]]
then 
    license-manager
else 
    go get -u github.com/oxequa/realize && realize start
    # go get github.com/derekparker/delve/cmd/dlv &&
    # go get github.com/pilu/fresh &&
    # /usr/bin/supervisord -n -c /etc/supervisor/conf.d/supervisord.conf
fi
