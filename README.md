#Cloud Rocker

Cloud Rocker is the convergence of [Cloud Foundry](http://cloudfoundry.org/index.html) and [Docker](https://www.docker.com/) aiming to bring the Cloud Foundry experience to a developer's machine. 

##Tutorial

This is a brief tutorial to ease you through your first rocking experience.

###Set up your Docker environment

Select either option 1 or option 2, then move to [Test your Docker environment](https://github.com/CloudCredo/cloudrocker#test-your-docker-environment).

####Option 1 : Easy - Mac/Windows/Linux - using Vagrant

```
vagrant up --provider virtualbox
vagrant ssh
```
Please use the Vagrantfile in the root of this repository. All later commands are entered as the vagrant user's shell in the vagrant VM.

Modify the [Vagrantfile](https://github.com/CloudCredo/cloudrocker/blob/master/Vagrantfile#L12) to mount your local workspace to make development easy.

####Option 2 : Advanced users - Linux only - using a local Docker daemon

```
$ go get github.com/cloudcredo/cloudrocker/rock
```
The user for 'rock' commands must have permissions to access the Docker daemon.

###Test your Docker environment

```$ rock docker```

You should see output similar to this:
```
Client API version: 1.15
Go version (client): go1.3
OS/Arch (client): linux/amd64
Server version: 1.3.0
Server API version: 1.15
Go version (server): go1.3.3
Git commit (server): c78088f
```

###Rock your local machine

Download the base Cloud Foundry container image.

```$ rock this```

###Add a buildpack 

```$ rock add-buildpack https://github.com/cloudfoundry/java-buildpack``` 

###Deploy your application

Change PWD to the sample Java application.

```$ cd /vagrant/sample-apps/java/```
  
Start the application.

```$ rock up```

```
Starting the CloudRocker container...
Running Buildpacks...
-----> Java Buildpack Version: 120c640
*--Buildpack output omitted--*
Started the CloudRocker container.
Deleting the CloudRocker container...
cloudrocker-staging
Deleted container.
Starting the CloudRocker container...
5b69950f351d2c843fe2ffd531edd87c09f19a368241ae37a6d2e025000dd6c8
Started the CloudRocker container.
Connect to your running application at http://localhost:8080/
```

You should now be able to browse the output on the vagrant machine.

```$ curl localhost:8080```

You should also be able to browse the output on your host machine.

[The rocking site.](http://localhost:8080/)

Please note the unsubtle [CloudCredo](http://www.cloudcredo.com/) advertising.
  
###Shut the application down

```$ rock off```

##Buildpacks

A great list of Cloud Foundry buildpacks is [available on the Cloud Foundry community wiki](https://github.com/cloudfoundry-community/cf-docs-contrib/wiki/Buildpacks). 

List buildpacks

```$ rock buildpacks```

```
cf-buildpack-php
go-buildpack
java-buildpack
nodejs-buildpack
python-buildpack
ruby-buildpack
```

Add a buildpack

```$ rock add-buildpack https://github.com/cloudfoundry/php-buildpack```

Remove a buildpack

```$ rock delete-buildpack php-buildpack```

Sample applications to use with the buildpacks are in [sample-apps](https://github.com/CloudCredo/cloudrocker/tree/master/sample-apps).

##Docker Images

###Building a Docker image from your application code

Ensure you have a corresponding buildpack installed for your application type.

```$ rock buildpacks```

Ensure your PWD is your application directory.

```$ cd <app_dir>```

Build a Docker image.

```$ rock build```

or with a tag

```$ rock build user/image:tag```
eg
```$ rock build hatofmonkeys/rocker-test:latest```


```
<output truncated>
Step 10 : CMD ["/bin/bash", "/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", "/app", "bundle", "exec", "rackup", "config.
ru", "-p", "$PORT"]
 ---> Running in fd810cea3db4
 ---> 4a88ad7d67ae
Removing intermediate container fd810cea3db4
Successfully built 4a88ad7d67ae
Created image.
```

In the example above the created image ID is 4a88ad7d67ae.

```$ docker images```

```
REPOSITORY                TAG                 IMAGE ID            CREATED             VIRTUAL SIZE
hatofmonkeys/rocker-test  latest              4a88ad7d67ae        13 minutes ago      609 MB
```

This image can be run like any other Docker image.

```
$ docker run -P -d hatofmonkeys/rocker-test:latest
0e4825d049ba5390699625be145a8d029a5e2899e52c8c5f967d35e08412f3ba
$ docker ps
CONTAINER ID        IMAGE                            COMMAND                CREATED             STATUS              PORTS                     NAMES
0e4825d049ba        hatofmonkeys/rocker-test:latest  /bin/bash /app/cloud   4 minutes ago       Up 4 minutes        0.0.0.0:49154->8080/tcp   sick_lumiere        
$ curl localhost:49154
Hello world!
```

The image can be uploaded to a Docker registry, or deployed and run in a system such as [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes), [Decker](https://github.com/hatofmonkeys/decker-release), or [Diego](http://thenewstack.io/docker-on-diego-cloud-foundrys-new-elastic-runtime/).

##External Services
  
Services can be connected to your application by adding a *vcap_services.json* file to the root directory of your application. This is demonstrated in the [ruby-with-services sample application](https://github.com/CloudCredo/cloudrocker/tree/master/sample-apps/ruby-with-services).

#####Note - the following steps are only appropriate to users of the Vagrant image (tutorial option 1 above).

Change PWD to the sample ruby-with-services sample application.

```$ cd /vagrant/sample-apps/ruby-with-services```

If necessary, install a Ruby buildpack.

```$ rock add-buildpack https://github.com/cloudfoundry/cf-buildpack-ruby```

Start the application.

```$ rock up```

Set a value for a key.

[http://localhost:8080/set/hello/to/world](http://localhost:8080/set/hello/to/world)

Get the value.

[http://localhost:8080/get/hello](http://localhost:8080/get/hello)

##Working with [Lattice](http://lattice.cf)

In the root directory of your application, the workflow is three simple commands to build and deploy your own containers to Lattice.

*replace hatofmonkeys/rocker-test:latest with your own user/image:tag*

```$ rock build hatofmonkeys/rocker-test:latest```

```$ docker push hatofmonkeys/rocker-test:latest```

```$ ltc create rocker-test hatofmonkeys/rocker-test:latest```

*note this may appear to timeout on slow connections. Watch ```ltc status rocker-test``` until the application is running*

##Debugging your app, your buildpack, your staging process

Build/staging artefacts are placed in $CLOUDROCKER_HOME. By default this is $HOME/cloudrocker, eg. /home/vagrant/cloudrocker. This is a treasure trove of interesting information when debugging staging failures.

##Potential Uses

####For application development

Cloud Rocker gives a fast-feedback, production-like Cloud Foundry environment on a developer's machine. Make a change, *rock up*, rinse, repeat. When you're finished, just *rock off*.

####For buildpack development

Cloud Rocker gives a fast-feedback environment for iterating on buildpacks. Buildpacks are stored in $CLOUDROCKER_HOME(default ~/cloudrocker)/buildpacks and can be edited directly.

####For continuous integration

Deploying Cloud Rocker to your CI server means you can quickly get feedback about your Cloud Foundry applications. 

####For deploying applications to non-Cloud-Foundry environments

If you aren't fortunate enough to work in an organisation with access to a proper PaaS such as Cloud Foundry - you can still use Cloud Foundry's buildpacks, via Cloud Rocker, to build applications. You will need to host the containers in an IaaS+ offering, such as [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes) or [Shipyard](https://github.com/shipyard/shipyard). IaaS+ systems leave the responsibility with the user to *rock* themselves.

##Developing Cloud Rocker

Development is against the 'develop' branch, promoted to 'master' by [Concourse](http://concourse.ci/).

There is a public tracker for this project [here](https://www.pivotaltracker.com/projects/1119430).

##FAQ

#####Why have you built Cloud Rocker?

Cloud Foundry, at a high level, has two responsibilities for applications: *staging* - so they are ready to be run, and then - *running* them. [Decker](http://www.cloudcredo.com/decker-docker-cloud-foundry/) brought Cloud Foundry's *running* experience to container developers. Cloud Rocker brings the Cloud Foundry *staging* experience to application developers. By separating these responsibilities we can discuss the right way to build containers, and the right way to run them.

As a [Cloud Foundry Community Advisory Board](http://cloudfoundry.org/about/index.html) member - anything I can do to illustrate the great experience Cloud Foundry brings to developers is worthwhile. I think everybody benefits from fast-feedback and small batch sizes. Get code into a production-like environment as soon as possible. Let's iterate faster.

#####How is Cloud Rocker different to Slugbuilder, Building, Buildstep, etc?

Cloud Rocker uses the Cloud Foundry components as far as possible to provide a production-like Cloud Foundry environment to developers.

#####How is this different to [Decker](http://www.cloudcredo.com/decker-docker-cloud-foundry/)?

[Decker](http://www.cloudcredo.com/decker-docker-cloud-foundry/) is about running containers in a remote Cloud Foundry. Cloud Rocker is about running Cloud Foundry applications in local containers.

#####How is this different to [BOSH-Lite](https://github.com/cloudfoundry/bosh-lite)?

[BOSH-Lite](https://github.com/cloudfoundry/bosh-lite) is a fantastic development environment for BOSH developers. Cloud Rocker is for application developers.

#####How is this different to Micro Cloud Foundry?

MCF has been defunct since Cloud Foundry v1, and even back then it ate a considerable amount of resources. Cloud Rocker runs on any machine capable of running Docker, only consuming the resources necessary to stage and run the application. 

#####Just how 'production-like' is this?

Cloud Rocker attempts to use the same components as Cloud Foundry's 'Diego'. 

The base filesystem is exactly the same as a Cloud Foundry container.
  
Environment variables are currently handled in a different manner to CF. File locations and user mapping are also slightly different, which may cause subtle issues. Cloud Rocker is under development to bridge this gap.
  
#####How do I enter the running container to 'poke around' in the shell?

We will be automating this process, for now use [nsenter](https://github.com/jpetazzo/nsenter).

#####Why can't I use my boot2docker setup on win/mac, not vagrant?

Feel free to try Cloud Rocker with boot2docker, and good luck with the volume mounts. 

#####Can I use a non-default base container image?

By default the Cloud Foundry *lucid64* image is used. You can choose to download a different base container image using the $ROCKER_ROOTFS_URL environment variable. To use *cflinuxfs2* instead:
```$ ROCKER_ROOTFS_URL=https://s3.amazonaws.com/blob.cfblob.com/04a2c4fc-3287-4525-9110-3ab3d84230b8 rock this```

#####What about 'Error response from daemon: Conflict, The name cloudrocker-runtime is already assigned to 7a519360a3d3. You have to delete (or rename) that container to be able to assign cloudrocker-runtime to a container again.'?

This is what happens when good Cloud Rockers turn bad. Simply run:

```$ docker rm cloudrocker-runtime```

We will automate this when we have a better understanding of all the scenarios in which it occurs.

#####Did you only create this project so you could have fun making endless *double entendres* in the README?

No. I enjoyed the portmanteau too.

##A message from our sponsors.

If you've read this far you've demonstrated a remarkable level of tenacity. [CloudCredo](http://www.cloudcredo.com/) are recruiting, and would [like to hear from you](http://www.cloudcredo.com/contact-us/).
