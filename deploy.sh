#!/bin/bash
if [ "$TRAVIS_BRANCH" == "master" ]; then
  curl https://cli-assets.heroku.com/install.sh | sh
  docker login -u _ -p "$HEROKU_TOKEN" registry.heroku.com
  docker build -t registry.heroku.com/vivalavinyl-service/web .
  docker push registry.heroku.com/vivalavinyl-service/web 
  heroku container:release web -a vivalavinyl-service
fi