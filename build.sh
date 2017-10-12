yum install -y golang
pushd cropimage_src/cropimage/
export PKG_CONFIG_PATH=/usr/lib/pkgconfig
go get ./...
go build 
popd
cp cropimage_src/cropimage/cropimage . 
