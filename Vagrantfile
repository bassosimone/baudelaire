Vagrant.configure(2) do |config|
  config.vm.synced_folder ".", "/baudelaire"
  config.vm.define "jessie64" do |jessie64|
    jessie64.vm.box = "debian/contrib-jessie64"
    jessie64.vm.provider "virtualbox" do |v|
      v.memory = 2048
    end
    jessie64.vm.provision "shell", inline: <<-SHELL
      cd /baudelaire
      sudo apt-get install -y golang
      go build
    SHELL
  end
end
