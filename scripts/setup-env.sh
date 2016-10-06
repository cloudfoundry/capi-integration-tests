export API_ENDPOINT="api.bosh-lite.com"
export API_DOMAIN="bosh-lite.com"
export SPACE="test"
export SERVICE="some-service"
export SYSLOG_DRAIN_SERVICE="log-drain-service"

export APP="env-vars-routes-service-bindings"
export APP_WITH_ENV_VARS="env-vars-routes-service-bindings"
export APP_WITH_MULTIPLE_ROUTES="env-vars-routes-service-bindings"
export APP_WITH_SERVICE_BINDING="env-vars-routes-service-bindings"
export APP_WITH_SYSLOG_DRAIN_URL="app-with-syslog-drain"
export UNSTAGED_APP='unstaged'
export STOPPED_APP='stopped'
export BUILDPACK_APP_TO_REPUSH="buildpack-app-to-repush"
export DIEGO_APP="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_ENV_VARS="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_MULTIPLE_ROUTES="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_SERVICE_BINDING="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_SYSLOG_DRAIN_URL="diego-app-with-syslog-drain"
export DIEGO_APP_WITH_MULTIPLE_PORTS="diego-app-with-multiple-ports"
export DIEGO_BUILDPACK_APP_TO_REPUSH="diego-buildpack-app-to-repush"
export DIEGO_DOCKER_APP_TO_REPUSH="diego-docker-app-to-repush"
export DIEGO_UNSTAGED_APP='diego-unstaged'
export DIEGO_STOPPED_APP='diego-stopped'
export APP_TO_RESTART_AND_RESTAGE_WITH_V3="v3-restart-restage"
export BUILDPACK_APP_TO_REPUSH_WITH_V3="buildpack-v3-repush"
export DOCKER_APP_TO_REPUSH_WITH_V3="docker-v3-repush"
export TASK_APP="task-app"
export UNSTAGED_APP_TO_STAGE_AND_START_WITH_V3='unstaged-for-v3'
export STOPPED_APP_TO_START_WITH_V3='stopped-for-v3'
export V3_APP="pushed-with-v3"
export JAVA_APP="java-app-start-command"
export NODE_APP="node-app-start-command"
export GOLANG_APP="golang-app-start-command"
export PHP_APP="php-app-start-command"
export PYTHON_APP="python-app-start-command"

cf api $API_ENDPOINT --skip-ssl-validation
cf auth admin admin
cf create-org test
cf create-space -o test $SPACE
cf target -o test -s $SPACE

cf enable-feature-flag task_creation
cf enable-feature-flag diego_docker
