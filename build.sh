yum install -y golang
cd cropimage_src
PKG_CONFIG_PATH=/usr/lib/pkg-config make
cd ..
cp cropimage_src/build/bin/cropimage . 
