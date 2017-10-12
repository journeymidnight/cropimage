FROM openresty/openresty:centos

WORKDIR image-crop
RUN yum install -y redis pango-devel wqy-zenhei-fonts.noarch  wqy-microhei-fonts.noarch  google-droid-sans-fonts.noarch


ADD libuuidx.so /usr/local/openresty/lualib/resty/
ADD uuid.lua /usr/local/openresty/lualib/resty/
ADD nginx.conf /usr/local/openresty/nginx/conf/
ADD puremagic.lua /usr/local/openresty/lualib/resty/



ADD cropimage /image-crop
ADD preinstall.sh /image-crop
ADD vips-8.3.1.tar.gz /image-crop
RUN sh preinstall.sh
RUN mkdir /var/log/redis /var/log/nginx
ADD image_processing.lua /image-crop


RUN yum install -y epel-release &&  yum install -y redis
ADD entrypoint.sh /image-crop
CMD ["/bin/bash", "entrypoint.sh"]
