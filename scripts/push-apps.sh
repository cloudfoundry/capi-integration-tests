#!/usr/bin/env bash

set -x

app_directory="assets/dora"
service_broker_directory="assets/service_broker"

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

function push_v3_app() {
  declare app_name=$1 args=$2

  pushd $app_directory
    cf v3-push $app_name $args
  popd
}

function push_service_broker() {
  pushd $service_broker_directory
    ./setup_new_broker.rb
  popd
}

function add_multiple_routes() {
  declare app_name=$1
  declare app_route1=$1"-route1"
  declare app_route2=$1"-route2"

  cf map-route $app_name bosh-lite.com --hostname $app_route1
  cf map-route $app_name bosh-lite.com --hostname $app_route2
}

function create_managed_service_instance() {
  declare service_instance_name=$1

  cf create-service "fake-service" "fake-plan" $service_instance_name
}

function set_env_vars_and_restage() {
  declare app_name=$1

  cf set-env $app_name "kittens" "cutest"
  cf restage $app_name
}

function bind_service_to_app_and_restage() {
  declare app_name=$1 service_name=$2

  cf bind-service $app_name $service_name
  cf restage $app_name
}

function push_app_with_syslog_drain() {
  declare app_name=$1
  declare syslog_drain_service=$2
  declare diego=$3

  if [$diego]; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi

  cf create-user-provided-service $syslog_drain_service -l logit.io/drain/here
  cf bind-service $app_name $syslog_drain_service

  cf restart $app_name
}

function main() {
  # push_service_broker
  # create_managed_service_instance $SERVICE

  # push_dea_app $APP_NAME
  # push_diego_app $DIEGO_APP_NAME

  # add_multiple_routes $DIEGO_APP_WITH_MULTIPLE_ROUTES_NAME
  # add_multiple_routes $APP_WITH_MULTIPLE_ROUTES_NAME

  # set_env_vars_and_restage $DIEGO_APP_WITH_ENV_VARS
  # set_env_vars_and_restage $APP_WITH_ENV_VARS

  # bind_service_to_app_and_restage $DIEGO_APP_WITH_SERVICE_BINDING_NAME $SERVICE
  # bind_service_to_app_and_restage $APP_WITH_SERVICE_BINDING_NAME $SERVICE

  push_app_with_syslog_drain $DIEGO_APP_WITH_SYSLOG_DRAIN_URL_NAME $SYSLOG_DRAIN_SERVICE true
  push_app_with_syslog_drain $APP_WITH_SYSLOG_DRAIN_URL_NAME $SYSLOG_DRAIN_SERVICE false

  # push_diego_app $DIEGO_BUILDPACK_APP_TO_REPUSH true
  # push_dea_app $BUILDPACK_APP_TO_REPUSH false

  # push_diego_app $DIEGO_DOCKER_APP_TO_REPUSH "-o cloudfoundry/diego-docker-app:latest" true
  # push_dea_app $DOCKER_APP_TO_REPUSH "-o cloudfoundry/diego-docker-app:latest" false

  # push_v3_app $V3_APP
  # push_v3_app $TASK_APP
  # push_v3_app $BUILDPACK_V3_APP_TO_REPUSH
  # push_v3_app $DOCKER_V3_APP_TO_REPUSH "-di cloudfoundry/diego-docker-app:latest"
}

source scripts/setup-env.sh
main
