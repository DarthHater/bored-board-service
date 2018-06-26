#!/bin/bash
docker login -u _ -p "$HEROKU_TOKEN" registry.heroku.com
docker build -t registry.heroku.com/vivalavinyl/vivalavinyl .
docker push registry.heroku.com/vivalavinyl/vivalavinyl docker login -u _ -p $HEROKU_TOKEN registry.heroku.com
