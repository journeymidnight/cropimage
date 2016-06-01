#!/bin/bash
rm -f *.png
#578*155
curl http://s3.lecloud.com/test/summary_bucket.png -o origin.png

#test0 /0/w/<LongEdge>/h/<ShortEdge> 指定单边缩略
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/0/w/90 -o mode0_1.png
result=`md5sum mode0_1.png | awk '{print $1}'`
if [ $result == "09b33446f289cb1eb6673887b6e390ff" ];then
	echo "success test0-1"
else
	echo "failed test0-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/0/h/80 -o mode0_2.png
result=`md5sum mode0_2.png | awk '{print $1}'`
if [ $result == "57e6fc37d7ff6c72a7ac3754d8a64ed2" ];then
	echo "success test0-2"
else
	echo "failed test0-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/0/h/300/l/1 -o mode0_3.png
result=`md5sum mode0_3.png | awk '{print $1}'`
if [ $result == "1c719435320cd61b6633925f583e6d8c" ];then
	echo "success test0-3"
else
	echo "failed test0-1"
fi


#test01 /1/w/<LongEdge>/h/<ShortEdge> 指定宽高缩略
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/1/w/90/h/100/e/1 -o mode1_1.png
result=`md5sum mode1_1.png | awk '{print $1}'`
if [ $result == "c8455aea23264f35d7435b715cbde4b7" ];then
	echo "success test1-1"
else
	echo "failed test1-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/1/w/90/h/100/e/0 -o mode1_2.png
result=`md5sum mode1_2.png | awk '{print $1}'`
if [ $result == "09b33446f289cb1eb6673887b6e390ff" ];then
	echo "success test1-2"
else
	echo "failed test1-2"
fi

#test02 /2/w/<LongEdge>/h/<ShortEdge>/l 强制宽高缩略
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/2/w/300/h/60/l/0 -o mode2_1.png
result=`md5sum mode2_1.png | awk '{print $1}'`
if [ $result == "8169c2e33865e994ca5d434ce6730c13" ];then
	echo "success test2-1"
else
	echo "failed test2-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/2/w/500/h/600/l/1 -o mode2_2.png
result=`md5sum mode2_2.png | awk '{print $1}'`
if [ $result == "20393604cf0d2b6e1b7933bf58128ed5" ];then
	echo "success test2-2"
else
	echo "failed test2-2"
fi

#test03 /3/p/1-1000 按比例缩放
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/3/p/50 -o mode3_1.png
result=`md5sum mode3_1.png | awk '{print $1}'`
if [ $result == "ebf16924867d89f5c2fac66dcaf8ed34" ];then
	echo "success test3-1"
else
	echo "failed test3-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/3/p/600 -o mode3_2.png
result=`md5sum mode3_2.png | awk '{print $1}'`
if [ $result == "49837c4653dbd451e3b355eac78d4a72" ];then
	echo "success test3-2"
else
	echo "failed test3-2"
fi

#test04 /4/w/50/h/100 缩略后填充,黑色
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/4/w/50/h/100 -o mode4_1.png
result=`md5sum mode4_1.png | awk '{print $1}'`
if [ $result == "0b82f5f7c36473ca283182ff39d94ed7" ];then
	echo "success test4-1"
else
	echo "failed test4-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/4/w/600/h/700 -o mode4_2.png
result=`md5sum mode4_2.png | awk '{print $1}'`
if [ $result == "f2c9cab8a8109edbd7a249bfec8dc7d4" ];then
	echo "success test4-2"
else
	echo "failed test4-2"
fi

#test05 /5/w/50/h/100 自动裁剪
curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/5/w/333/h/100 -o mode5_1.png
result=`md5sum mode5_1.png | awk '{print $1}'`
if [ $result == "d9cff1c10689981765b416038a3c1b34" ];then
	echo "success test5-1"
else
	echo "failed test5-1"
fi

curl http://127.0.0.1:9000/s3.lecloud.com/test/summary_bucket.png?imageview/5/w/800/h/300/l/1 -o mode5_2.png
result=`md5sum mode5_2.png | awk '{print $1}'`
if [ $result == "243ced923d1cee20d8abfde076a992f6" ];then
	echo "success test5-2"
else
	echo "failed test5-2"
fi

#test watermark watermark/1/text/<encodeText>/font/<Font>/color/<Color>/size/<Size>/o/<Opacity>/w/<width>/x/<distanceX>/y/<distanceY>
curl http://127.0.0.1:9000/s3.lecloud.com/image-test/image112875085.jpg?watermark/1/text/5LmQ6KeG5LqR5Zu-54mH5pyN5Yqh/color/0000FF/size/100/o/99/w/500/x/-10/y/-10 -o watermark1.1.jpg
result=`md5sum watermark1.1.jpg | awk '{print $1}'`
if [ $result == "e8316243b2cb1d5009371552170ac4ef" ];then
	echo "success watermark-1-1"
else
	echo "failed watermark-1-1"
fi
