#!/bin/bash
yum install pango-devel -y
sh ./preinstall.sh
yum install http://s3.lecloud.com/test/openresty-1.9.7.4-1.el6.x86_64.rpm -y
yum install redis -y
yum install monit -y
cp ./libuuidx.so /usr/local/openresty/lualib/resty/
cp ./uuid.lua /usr/local/openresty/lualib/resty/
cp ./puremagic.lua /usr/local/openresty/lualib/resty/
cp .//monit.conf /etc
mkdir /var/log/redis
mkdir /var/log/nginx
yum install wqy-zenhei-fonts.noarch -y
yum install wqy-microhei-fonts.noarch -y
yum install google-droid-sans-fonts.noarch -y
cp ./logrotate.d/* /etc/logrotate.d/
/etc/init.d/monit start
/etc/init.d/crond start
