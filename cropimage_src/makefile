.PHONY:cropimage

#GOPATH = /home/zhangdongmao/go/:$(PWD)/build:$(PWD)/Godeps/_workspace
GOPATH = $(PWD)/build:$(PWD)/Godeps/_workspace
export GOPATH
URL = github.com/thesues
REPO = cropimage
URLPATH = $(PWD)/build/src/$(URL)

cropimage:
	@[ -d $(URLPATH) ] || mkdir -p $(URLPATH)
	@ln -nsf $(PWD) $(URLPATH)/$(REPO)
	go install $(URL)/$(REPO)/cropimage

install:
	@[ -d $(LOGPATH) ]|| mkdir -p $(LOGPATH)
	@[ -d $(PIDPATH) ]|| mkdir -p $(PIDPATH)
	install -D build/bin/cropimage $$DESTDIR/usr/bin/cropimage

clean:
	rm -fr rpm-build
	rm -rf build
	rm -rf *.rpm

rpm:
	sh package/rpmbuild.sh
