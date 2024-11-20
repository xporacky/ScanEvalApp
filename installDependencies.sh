#!/bin/bash
add_to_bashrc=false
sudo apt update
sudo apt install -y texlive-full xxd wget gcc g++ unzip sqlite3 tesseract-ocr tesseract-ocr-slk 
sudo apt install -y build-essential cmake git pkg-config libgtk-3-dev libavcodec-dev libavformat-dev libswscale-dev libv4l-dev libxvidcore-dev libx264-dev libjpeg-dev libpng-dev libtiff-dev gfortran openexr libatlas-base-dev python3-dev python3-numpy libtbbmalloc2 libtbb-dev
# go instalation
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
#install opencv
cd /tmp
# Download and unzip OpenCV
wget -O opencv.zip https://github.com/opencv/opencv/archive/4.10.0.zip
unzip opencv.zip
# Create build directory
mkdir -p build && cd build
 
# Configure
cmake  ../opencv-4.10.0
 
# Build
cmake --build .

#goCV installation
cd /usr/local/go/src && sudo git clone https://github.com/hybridgroup/gocv.git
cd gocv
make install
go run ./cmd/version/main.go

#adding go to path
if [ "$adding" = true ] ; then
      echo "Adding to bashrc"
      echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
      exec $SHELL 
fi

go version