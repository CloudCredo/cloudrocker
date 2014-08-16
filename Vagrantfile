# Thanks to Phusion for the great boxes
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.network "forwarded_port", guest: 8080, host: 8080

#comment out the two lines below, and uncomment the following block, to change the VM
  config.vm.box = "cloudfocker-0.0.1-amd64"
  config.vm.box_url = "https://s3.amazonaws.com/cloudfocker/vagrantboxes/cloudfocker-0.0.1-vbox.box"

=begin
  config.vm.box = "phusion-open-ubuntu-14.04-amd64"
  config.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-14.04-amd64-vbox.box"
  # Or, for Ubuntu 12.04:
  #config.vm.box = "phusion-open-ubuntu-12.04-amd64"
  #config.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-12.04-amd64-vbox.box"

  config.vm.provider :vmware_workstation do |f, override|
    override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-14.04-amd64-vmwarefusion.box"
    #override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-12.04-amd64-vmwarefusion.box"
  end

  config.vm.provider :vmware_fusion do |f, override|
    override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-14.04-amd64-vmwarefusion.box"
    #override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-12.04-amd64-vmwarefusion.box"
  end

  if Dir.glob("#{File.dirname(__FILE__)}/.vagrant/machines/default/*/id").empty?
    # Install Docker
    pkg_cmd = "wget -q -O - https://get.docker.io/gpg | apt-key add -;" \
      "echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list;" \
      "apt-get update -qq; apt-get install -q -y --force-yes lxc-docker; "
    # Add vagrant user to the docker group
    pkg_cmd << "usermod -a -G docker vagrant; "

    # Install redis for services demo
    pkg_cmd << "apt-get install -q -y --force-yes redis-server; "
    pkg_cmd << "sed -i 's/bind 127.0.0.1/#bind 127.0.0.1/g' /etc/redis/redis.conf; "
    pkg_cmd << "/etc/init.d/redis-server restart; "

    # Install golang + focker
    pkg_cmd << "wget -q -O /tmp/go.tgz http://golang.org/dl/go1.3.linux-amd64.tar.gz; "
    pkg_cmd << "tar xzf /tmp/go.tgz -C /usr/lib; "
    pkg_cmd << "apt-get install -q -y --force-yes bzr mercurial; "
    pkg_cmd << "mkdir -p /home/vagrant/go; chown vagrant /home/vagrant/go; "
    pkg_cmd << "echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.bashrc; "
    pkg_cmd << "echo 'export GOROOT=/usr/lib/go' >> /home/vagrant/.bashrc; "
    pkg_cmd << "echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> /home/vagrant/.bashrc; "
    pkg_cmd << "cd /home/vagrant/go; su vagrant -c 'export GOROOT=/usr/lib/go; export GOPATH=/home/vagrant/go; /usr/lib/go/bin/go get github.com/tools/godep; /usr/lib/go/bin/go get -v github.com/cloudcredo/cloudfocker/fock; cd /home/vagrant/go/src/github.com/cloudcredo/cloudfocker/fock; PATH=$PATH:/usr/lib/go/bin /home/vagrant/go/bin/godep restore; /usr/lib/go/bin/go install'"

    config.vm.provision :shell, :inline => pkg_cmd
  end
=end
end