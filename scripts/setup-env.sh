export API_ENDPOINT="api.bosh-lite.com"
export API_DOMAIN="bosh-lite.com"
export SPACE="test"
export APP="env-vars-routes-service-bindings"
export APP_WITH_ENV_VARS="env-vars-routes-service-bindings"
export APP_WITH_MULTIPLE_ROUTES="env-vars-routes-service-bindings"
export APP_WITH_SERVICE_BINDING="env-vars-routes-service-bindings"
export APP_WITH_SYSLOG_DRAIN_URL="app-with-syslog-drain"
export BUILDPACK_APP_TO_REPUSH="buildpack-app-to-repush"
export DOCKER_APP_TO_REPUSH="docker-app-to-repush"
export DIEGO_APP="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_ENV_VARS="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_MULTIPLE_ROUTES="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_SERVICE_BINDING="diego-env-vars-routes-service-bindings"
export DIEGO_APP_WITH_SYSLOG_DRAIN_URL="diego-app-with-syslog-drain"
export DIEGO_BUILDPACK_APP_TO_REPUSH="diego-buildpack-app-to-repush"
export DIEGO_DOCKER_APP_TO_REPUSH="diego-docker-app-to-repush"
export APP_TO_RESTART_AND_RESTAGE_WITH_V3="v3-restart-restage"
export BUILDPACK_APP_TO_REPUSH_WITH_V3="buildpack-v3-repush"
export DOCKER_APP_TO_REPUSH_WITH_V3="docker-v3-repush"
export TASK_APP="task-app"
export SERVICE="some-service"
export SYSLOG_DRAIN_SERVICE="log-drain-service"

cf api $API_ENDPOINT --skip-ssl-validation
cf auth admin admin
cf create-org test
cf create-space -o test $SPACE
cf target -o test -s $SPACE

cf enable-feature-flag task_creation
cf enable-feature-flag diego_docker
