#压力测试

##安装wrk

##运行

./wrk -t100 -c100 -d30s -T30s --script=../image_crop/test/stress.lua --latency http://impress.lecloud.com
