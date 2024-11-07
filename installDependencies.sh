#!/bin/bash
add_to_bashrc=false
sudo apt install -y texlive-full xxd wget
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz
sudo rm go1.23.2.linux-amd64.tar.gz
adding=true
while getopts "f" opt; do
  case $opt in
    f)
      addingx=false 
      echo "Not adding to bashrc"
      ;;
  esac
done
if [ "$adding" = true ] ; then
      echo "Adding to bashrc"
      echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
      exec $SHELL 
fi

go version
