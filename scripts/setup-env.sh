export API_ENDPOINT="api.bosh-lite.com"
export API_DOMAIN="bosh-lite.com"
export SPACE="test"
export APP_NAME="meow"
export APP_WITH_ENV_VARS="meow"
export APP_WITH_MULTIPLE_ROUTES_NAME="meow"
export APP_WITH_SERVICE_BINDING_NAME="meow"
export APP_WITH_SYSLOG_DRAIN_URL_NAME="log-drain-app"
export BUILDPACK_APP_TO_REPUSH="repush-buildpack-app"
export DOCKER_APP_TO_REPUSH="repush-docker-app"
export DIEGO_APP_NAME="diego-meow"
export DIEGO_APP_WITH_ENV_VARS="diego-meow"
export DIEGO_APP_WITH_MULTIPLE_ROUTES_NAME="diego-meow"
export DIEGO_APP_WITH_SERVICE_BINDING_NAME="diego-meow"
export DIEGO_APP_WITH_SYSLOG_DRAIN_URL_NAME="diego-log-drain-app"
export DIEGO_BUILDPACK_APP_TO_REPUSH="diego-repush-buildpack-app"
export DIEGO_DOCKER_APP_TO_REPUSH="diego-repush-docker-app"
export V3_APP="v3-app"
export BUILDPACK_V3_APP_TO_REPUSH="repush-v3-buildpack-app"
export DOCKER_V3_APP_TO_REPUSH="repush-v3-docker-app"
export TASK_APP="task-app"
export SERVICE="some-db"
export SYSLOG_DRAIN_SERVICE="log-drain-service"

cf api $API_ENDPOINT --skip-ssl-validation
cf auth admin admin
cf create-org test
cf create-space -o test $SPACE
cf target -o test -s $SPACE

cf enable-feature-flag task_creation
cf enable-feature-flag diego_docker
