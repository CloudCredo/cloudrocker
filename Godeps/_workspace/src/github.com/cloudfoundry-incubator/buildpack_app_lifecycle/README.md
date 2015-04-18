# buildpack app lifecycle 

The buildpack lifecycle implements the traditional Cloud Foundry deployment
strategy.

The **Builder** downloads buildpacks and app bits, and produces a droplet.

The **Launcher** runs the start command using a standard rootfs and
environment.

The **Healthcheck** runs a tcp port check, defaulting to port 8080.

Read about the app lifecycle spec here: https://github.com/cloudfoundry-incubator/diego-design-notes#app-lifecycles
