#!/usr/bin/env bash

app_directory="assets/dora"

function push_dea_app() {
  declare app_name=$1 args=$2
  pushd $app_directory
    cf push $app_name

  popd
}

function push_diego_app() {
  declare app_name=$1 args=$2
  pushd $app_directory
    cf push $app_name --no-start
    cf enable-diego $app_name
    cf start $app_name
  popd
}

function push_diego_app_with_env_vars() {
  declare app_name=$1

  push_diego_app $app_name
  cf set-env $app_name "kittens" "cutest"
  cf restage $app_name
}

function push_diego_app_with_service_binding() {
  declare app_name=$1

  push_diego_app $app_name
  cf bind-service $app_name "oink"
  cf restage $app_name
}

function push_diego_app_with_multiple_routes() {
  declare app_name=$1
  declare app_route1=$1 + "-route1"
  declare app_route2=$1 + "-route2"
  declare app_route3=$1 + "-route3"

  push_diego_app $app_name

  cf map-route $app_name bosh-lite.com --hostname $app_route1
  cf map-route $app_name bosh-lite.com --hostname $app_route2
  cf map-route $app_name bosh-lite.com --hostname $app_route3
}

function push_diego_app_with_syslog_drain() {
  declare app_name=$1

  push_diego_app $app_name

  cf create-user-provided-service log-drain-service -l logit.io/drain/here
  cf bind-service $app_name log-drain-service
}

function main() {
  push_diego_app $DIEGO_APP_NAME
  push_diego_app_with_env_vars $DIEGO_APP_WITH_ENV_VARS
  push_diego_app_with_service_binding $DIEGO_APP_WITH_SERVICE_BINDING_NAME
  push_diego_app_with_multiple_routes $DIEGO_APP_WITH_MULTIPLE_ROUTES_NAME
  push_diego_app_with_syslog_drain $DIEGO_APP_WITH_SYSLOG_DRAIN_URL_NAME
  push_diego_app $DIEGO_BUILDPACK_APP_TO_REPUSH
  push_diego_app $DIEGO_DOCKER_APP_TO_REPUSH -o cloudfoundry/diego-docker-app:latest
}

main
