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

function push_app_with_env_vars() {
  declare app_name=$1
  declare diego=$2

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi

  cf set-env $app_name "kittens" "cutest"
  cf restage $app_name
}

function push_app_with_env_vars() {
  declare app_name=$1
  declare diego=$2

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi

  cf set-env $app_name "kittens" "cutest"
  cf restage $app_name
}

function push_app_with_service_binding() {
  declare app_name=$1
  declare diego=$2

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi
  
  cf bind-service $app_name "oink"
  cf restage $app_name
}

function push_app_with_multiple_routes() {
  declare app_name=$1
  declare app_route1=$1"-route1"
  declare app_route2=$1"-route2"
  declare diego=$2

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi


  cf map-route $app_name bosh-lite.com --hostname $app_route1
  cf map-route $app_name bosh-lite.com --hostname $app_route2
}

function push_app_with_syslog_drain() {
  declare app_name=$1
  declare diego=$2

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi

  cf create-user-provided-service log-drain-service -l logit.io/drain/here
  cf bind-service $app_name log-drain-service
}

function main() {
  push_dea_app $APP_NAME
  push_diego_app $DIEGO_APP_NAME

  push_app_with_env_vars $DIEGO_APP_WITH_ENV_VARS true
  push_app_with_env_vars $APP_WITH_ENV_VARS false

  push_app_with_service_binding $DIEGO_APP_WITH_SERVICE_BINDING_NAME true
  push_app_with_service_binding $APP_WITH_SERVICE_BINDING_NAME false

  push_app_with_multiple_routes $DIEGO_APP_WITH_MULTIPLE_ROUTES_NAME true
  push_app_with_multiple_routes $APP_WITH_MULTIPLE_ROUTES_NAME false

  push_app_with_syslog_drain $DIEGO_APP_WITH_SYSLOG_DRAIN_URL_NAME true
  push_app_with_syslog_drain $APP_WITH_SYSLOG_DRAIN_URL_NAME false

  push_diego_app $DIEGO_BUILDPACK_APP_TO_REPUSH true
  push_dea_app $BUILDPACK_APP_TO_REPUSH false

  push_diego_app $DIEGO_DOCKER_APP_TO_REPUSH -o cloudfoundry/diego-docker-app:latest true
  push_dea_app $DOCKER_APP_TO_REPUSH -o cloudfoundry/diego-docker-app:latest false
}

main
