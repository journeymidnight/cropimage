FROM thesues/centos7-vips

WORKDIR /image-crop

COPY ["puremagic.lua", "libuuidx.so", "uuid.lua" , "/usr/local/openresty/lualib/resty/"]

COPY ["nginx.conf" , "/usr/local/openresty/nginx/conf/"]


RUN mkdir -p /var/log/redis /var/log/nginx


COPY ["cropimage", "image_processing.lua", "entrypoint.sh", "/image-crop/"]
CMD ["/bin/bash", "entrypoint.sh"]
