# Image processing

## Url rules


Image processing serviced could be accessed by HTTP's GET Interface

All the image processing parameters are encoded in URL


the generic format of image processing request is as below:


```
    Request:

    http://impress.lecloud.com/<URL>@<PARAMETERS>
    URL: any source image URL
    PARAMETERS: parameters to control the processing

    Response:
 
    HTTP/1.1 200 OK
    Content-Type: <ImageMimeType>

    <ImageBinaryData>

```


Example:


```
    http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/0/w/100
```

Supported File Format: jpg, png, webp, tiff

Note:The order of all parameters must be as the same as this document's order. Optional parameters can be empty


## Image Resize and Crop

Parameters:

```
    imageview/<mode>/w/<width>
                /h/<hight>                
                /e/<edge>
                /l/<enlarger>
                /p/<proportion>
```

6 different modes are supported:


mode list:


| Value         | Description                                      | options  | others    |
|:------------- |:------------------------------                   |:-------- |:----------|
| 0             | either width or height is fixed, fixed w/h ratio | w,h,l    | can only use width or height, not both
| 1             | resize, fixed width and height, fixed w/h ratio  | w,h,e,l  | choose either long edge or short edge as resize base. the resized image may exceed.
| 2             | resize, fixed width and height                   | w,h,l    | both width and height is a must.
| 3             | resize on proportion, fixed w/h ratio            | p        | 
| 4             | fill gap after resize, fixed w/h ratio           | w,h      | use color black to fill the gap after resize
| 5             | auto resize, fixed w/h ratio                     | w,h,l    | automally resize the image, if there is any exceeding, only crop the center part of the image 




optional parameter list:

| Name          | Description                                  | Range|
|:------------- |:-------------------------------------------  |:-----|
| w             | width of target image                        | 1-4096 |
| h             | heigth of target image                       | 1-4096 |
| e             | edge priority                                | 0: long edge priority, 1 short edge priority, default: 0|
| l             | if the target image is larger than original target, 1 means allowing this operation anyway. 0 means refuse the operation|0/1, default: 0|
| p             | proportion, 100 means the same size, proportion which is less than 100 means shrink, vice versa|1-1000 |

#####Examples



mode 0：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/0/w/90

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/0/h/80

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/0/h/300/l/1

mode 1：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/1/w/90/h/100/e/1

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/1/w/90/h/100/e/0

mode 2：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/2/w/300/h/60/l/0

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/2/w/500/h/600/l/1

mode 3：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/3/p/50

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/3/p/600

mode 4：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/4/w/50/h/100

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/4/w/600/h/700

mode 5：

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/5/w/333/h/100

http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@imageview/5/w/800/h/300/l/1



## Waterprint 

### Introduction 

The modes of waterprint have image waterprint, text waterprint, image/text hybrid waterprint.

For now, only text waterprint is supported.


Parameters:

```
    watermark/<mode>/o/<opacity>
                /w/<width>                
                /x/<distanceX>
                /y/<distanceY>

```

### waterprint mode

| Value         | Description                  |
|:------------- |:-----------------------------|
| 0             | image waterprint             | 
| 1             | text waterprint              |
| 2             | image/text hybrid waterprint | 


### waterprint parameters

| Value        | Description   |type |
|:------------- |:-------------|:-----|
| o      | Opacity, if a image waterprint, make the image opacity，if a text waterprint，make the text opacity <br>default:100， meaning 100% (opacity) Range: [0-100]         | optional | 
| w      | the total width of one line, default: 200px  | optional|
| x      | horizontal distance between waterprint and the orign image's border, positive value means relative to the left top, negtive value means relative to the right bottom <br>default: -10px |optional|
| y      | vertical distance  between waterprint and the orign image's border,  positive value means relative to the left top, negtive value means relative to the right bottom <br>default: -10px |optional|



### text waterprint

#### request URL

	watermark/1/text/<encodeText>/font/<Font>/color/<Color>/size/<Size>/o/<Opacity>/w/<width>/x/<distanceX>/y/<distanceY>

#### parameters

| name        | Description           | type |
|:------------- |:-------------|:-----|
| text      | waterprint content(encoding is required)          | optional |
| font      | waterprint fonts(encoding is required) <br>default: 文泉驿正黑   | optional |
| color     | color of waterprint  <br>default: 000000 (black) | optional|
| size      | font size of waterprint  <br>default:40  |  optional|

Notes：

* URL-safed base64 encoding for UTF-8 text
* Method:http://kjur.github.io/jsjws/tool_b64uenc.html (open source tools:https://www.npmjs.com/package/urlsafe-base64)
* The color must be encoded as hex character and upper case, such as WHITE(FFFFFF)

##### Supported Fonts


| fonts        | url-safe base64|
|:------------- | :-------------
| 文泉驿正黑      | 5paH5rOJ6am_5q2j6buR
| 文泉驿微米黑    | 5paH5rOJ6am_5b6u57Gz6buR
| Droid Sans Fallback     | RHJvaWQgU2FucyBGYWxsYmFjaw==

#####Example
http://impress.lecloud.com/s3.lecloud.com/image-test/cat.jpg@watermark/1/text/5LmQ6KeG5LqR5Zu-54mH5pyN5Yqh/font/5paH5rOJ6am_5b6u57Gz6buR/color/00EEFF/size/100/o/100/w/500/x/-10/y/-10



