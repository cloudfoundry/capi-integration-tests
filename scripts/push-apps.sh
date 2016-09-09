#!/usr/bin/env bash

set -x

app_directory="assets/dora"
service_broker_directory="assets/service_broker"

function push_dea_app() {
  declare app_name=$1 args=$2
  pushd $app_directory
    cf push $app_name $args
  popd
}

function push_diego_app() {
  declare app_name=$1 args=$2

  pushd $app_directory
    cf push $app_name --no-start ${args}
    cf enable-diego $app_name
    cf start $app_name
  popd
}

function push_unstaged_app() {
  declare app_name=$1 diego=$2

  pushd $app_directory
    cf push $app_name --no-start
  popd

  if $diego; then
    cf enable-diego $app_name
  fi
}

function stop_app() {
  declare app_name=$1

  cf stop $app_name
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

function create_route() {
  declare host=$1

  cf create-route $SPACE bosh-lite.com -n $host
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

  if $diego; then
    push_diego_app $app_name
  else
    push_dea_app $app_name
  fi

  cf create-user-provided-service $syslog_drain_service -l logit.io/drain/here
  cf bind-service $app_name $syslog_drain_service

  cf restart $app_name
}

function set_ports() {
  declare app_name=$1 ports=$2

  guid=$(cf app ${app_name} --guid)
  cf curl /v2/apps/${guid} -X PUT -d "{\"ports\": ${ports}}"
}

function add_route_with_port() {
  declare app_name=$1 route=$2 port=$3

  app_guid=$(cf app ${app_name} --guid)
  route_guid=$(cf curl /v2/routes?q=host:${route} | jq -r '.resources[0].metadata.guid')
  cf curl /v2/route_mappings -X POST -d "{\"app_port\": ${port}, \"app_guid\": \"${app_guid}\", \"route_guid\": \"${route_guid}\"}"
}

function main() {
  push_service_broker
  create_managed_service_instance $SERVICE

  push_dea_app $APP
  push_diego_app $DIEGO_APP

  add_multiple_routes $DIEGO_APP_WITH_MULTIPLE_ROUTES
  add_multiple_routes $APP_WITH_MULTIPLE_ROUTES

  set_env_vars_and_restage $DIEGO_APP_WITH_ENV_VARS
  set_env_vars_and_restage $APP_WITH_ENV_VARS

  bind_service_to_app_and_restage $DIEGO_APP_WITH_SERVICE_BINDING $SERVICE
  bind_service_to_app_and_restage $APP_WITH_SERVICE_BINDING $SERVICE

  push_app_with_syslog_drain $DIEGO_APP_WITH_SYSLOG_DRAIN_URL $SYSLOG_DRAIN_SERVICE true
  push_app_with_syslog_drain $APP_WITH_SYSLOG_DRAIN_URL $SYSLOG_DRAIN_SERVICE false

  push_diego_app $DIEGO_BUILDPACK_APP_TO_REPUSH
  push_dea_app $BUILDPACK_APP_TO_REPUSH

  push_diego_app $DIEGO_DOCKER_APP_TO_REPUSH "-o cloudfoundry/diego-docker-app:latest"
  push_dea_app $DOCKER_APP_TO_REPUSH "-o cloudfoundry/diego-docker-app:latest"

  push_diego_app $APP_TO_RESTART_AND_RESTAGE_WITH_V3
  push_diego_app $TASK_APP
  push_diego_app $BUILDPACK_APP_TO_REPUSH_WITH_V3
  push_diego_app $DOCKER_APP_TO_REPUSH_WITH_V3 "-o cloudfoundry/diego-docker-app:latest"

  push_unstaged_app $UNSTAGED_APP false
  push_unstaged_app $DIEGO_UNSTAGED_APP true
  push_unstaged_app $UNSTAGED_APP_TO_STAGE_AND_START_WITH_V3 true

  push_dea_app $STOPPED_APP
  push_diego_app $DIEGO_STOPPED_APP
  push_diego_app $STOPPED_APP_TO_START_WITH_V3
  stop_app $STOPPED_APP
  stop_app $DIEGO_STOPPED_APP
  stop_app $STOPPED_APP_TO_START_WITH_V3

  nothing_to_see_here=$app_directory
  app_directory="assets/lattice-app"

  push_diego_app $DIEGO_APP_WITH_MULTIPLE_PORTS '-u none --no-route'
  create_route $DIEGO_APP_WITH_MULTIPLE_PORTS
  set_ports $DIEGO_APP_WITH_MULTIPLE_PORTS "[9090, 9191]"
  add_route_with_port $DIEGO_APP_WITH_MULTIPLE_PORTS $DIEGO_APP_WITH_MULTIPLE_PORTS 9191

  app_directory=$nothing_to_see_here

  push_v3_app $V3_APP
}

source scripts/setup-env.sh
main
