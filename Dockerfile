FROM thesues/centos7-vips

WORKDIR /image-crop

RUN yum-config-manager --add-repo https://openresty.org/package/centos/openresty.repo && yum install -y openresty
RUN yum install -y redis wqy-zenhei-fonts.noarch  wqy-microhei-fonts.noarch  google-droid-sans-fonts.noarch


ADD libuuidx.so /usr/local/openresty/lualib/resty/
ADD uuid.lua /usr/local/openresty/lualib/resty/
ADD nginx.conf /usr/local/openresty/nginx/conf/
ADD puremagic.lua /usr/local/openresty/lualib/resty/



ADD cropimage /image-crop
RUN mkdir /var/log/redis /var/log/nginx
ADD image_processing.lua /image-crop


RUN yum install -y epel-release &&  yum install -y redis
ADD entrypoint.sh /image-crop
CMD ["/bin/bash", "entrypoint.sh"]
