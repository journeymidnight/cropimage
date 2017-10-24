FROM centos:7
WORKDIR /build
RUN yum install -y pango-devel wqy-zenhei-fonts.noarch  wqy-microhei-fonts.noarch  google-droid-sans-fonts.noarch;yum -y clean all 
ADD preinstall.sh /build
ADD vips-8.3.1.tar.gz /build
RUN bash preinstall.sh; yum -y clean all
