# How to Run These Tests

1. Deploy cf-release master
1. Run `scripts/push-apps.sh`
1. Check out the `migrate` branch of cloud_controller_ng
1. Deploy
1. `source scripts/setup-env.sh && CONFIG=$(PWD)/integration_config.json ginkgo migration
