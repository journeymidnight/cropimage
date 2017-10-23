package main


import (
        "fmt"
        "github.com/garyburd/redigo/redis"
        "github.com/thesues/bimg"
        "net/http"
        "encoding/json"
        "regexp"
        "strings"
        "errors"
        "strconv"
        "io/ioutil"
        "time"
        "log"
        "os"
        "io"
        "syscall"
        "github.com/urfave/cli"
        _ "net/http/pprof"
    "encoding/base64"
)


type TaskData struct {
    Uuid string    `json:"uuid"`
    Url string `json:"url"`
}

type CropTask struct {
        mode int
        width int
        height int
        edge int
        large int
    proportion int
}

type WaterMarkTask struct {
        mode int
        object string
        text string
        font string
        size string
        color bimg.Color
        order int
        align int
        interval int
        position int
        opacity float32 
        width int
        x int
        y int
}
        
type FinishTask struct {
        code int
        uuid string
        url string
        blob []byte
}

const (
        IMAGEVIEW = iota
        WATERMARK
)

var logger * log.Logger
var imageviewPattern = regexp.MustCompile("imageview/(?P<mode>[0-9])(?P<mwidth>/w/[0-9]+)?(?P<mheight>/h/[0-9]+)?(?P<medge>/e/[0-2])?(?P<mlarge>/l/[0-1])?(?P<mproportion>/p/[0-9]+)?")
var watermarkPattern = regexp.MustCompile("watermark/(?P<mode>[0-9])?(?P<mobject>/object/[a-zA-Z0-9-_=]+)?(?P<mtext>/text/[a-zA-Z0-9-_=]+)?(?P<mfont>/font/[a-zA-Z0-9-_=]+)?(?P<mcolor>/color/[0123456789ABCDEDFabcdef]{6})?(?P<msize>/size/[0-9]+)?(?P<mo>/o/[0-9]+)?(?P<mwidth>/w/[0-9]+)?(?P<mx>/x/[-0-9]+)?(?P<my>/y/[-0-9]+)?")
var imageviewNames = imageviewPattern.SubexpNames()
var watermarkNames = watermarkPattern.SubexpNames()

func returnError(code int, uuid string, url string, Q chan FinishTask) {
        Q <- FinishTask{code, uuid, url, nil}
}

func PreProcess(result map[string]string) (CropTask, error){

        n := CropTask{}

        if result["mode"] == "" {
                logger.Printf("no crop mode")
                return CropTask{}, errors.New("no crop data")
        }

        n.mode, _ = strconv.Atoi(result["mode"]) //convert to int

        if result["mwidth"] == "" {
                n.width = 0
        } else {
                n.width, _ = strconv.Atoi(strings.Split(result["mwidth"], "/")[2])
        }

        if result["mheight"] == ""{
                n.height = 0
        } else {
                n.height, _ = strconv.Atoi(strings.Split(result["mheight"], "/")[2])
        }

        if result["medge"] == ""{
                n.edge = 0
        } else {
                n.edge, _ = strconv.Atoi(strings.Split(result["medge"], "/")[2])
        }

        if result["mlarge"] == ""{
                n.large = 0
        } else {
                n.large, _ = strconv.Atoi(strings.Split(result["mlarge"], "/")[2])
        }

        if n.mode != 3 && n.width == 0 && n.height == 0 {
                return CropTask{}, errors.New("no width or height specified")
        } 

        if result["mproportion"] == ""{
                n.proportion = 100
        } else {
                n.proportion, _ = strconv.Atoi(strings.Split(result["mproportion"], "/")[2])
        }

        if n.mode == 0 {
                if n.width != 0 && n.height != 0 {
                        return CropTask{}, errors.New("can not specify width and hight both in mode 0")        
                }
        }

        if n.mode == 1 || n.mode == 2 {
                if n.width == 0 || n.height == 0 {
                        return CropTask{}, errors.New("must specify width and hight both in mode 1 and mode2")        
                }
        }

    if n.mode == 3 {
        if n.proportion < 1 || n.proportion > 1000 {
                        return CropTask{}, errors.New("proportion exceed limitation")
                }
    } 
        return n, nil
}

func base64UrlDecode(src string) (string, error) {
        str:= strings.Replace(src, "-", "+", -1)
        str = strings.Replace(str, "_", "/", -1)
        tmp, err := base64.StdEncoding.DecodeString(str)
        if err != nil {
                return "", err
        }
        return string(tmp), nil
}

func PreProcessWaterMark(result map[string]string) (WaterMarkTask, error){

        n := WaterMarkTask{}
        var tmp []byte
        var err error
        if result["mode"] == "" {
                logger.Printf("no watermark mode")
                return WaterMarkTask{}, errors.New("no watermark data")
        }

        n.mode, _ = strconv.Atoi(result["mode"]) //convert to int

        if result["mobject"] == "" {
                n.object = ""
                logger.Printf("mobject empty",n.object)
        } else {
                n.object, err = base64UrlDecode(strings.Split(result["mobject"], "/")[2])
                if err != nil {
                        return WaterMarkTask{}, errors.New("mobject decode error")
                }
        }

        if result["mtext"] == "" {
                logger.Printf("mtext empty",n.text)
                n.text = ""
        } else {
                n.text, err = base64UrlDecode(strings.Split(result["mtext"], "/")[2])
                if err != nil {
                        return WaterMarkTask{}, errors.New("mtext decode error")
                }
                logger.Printf("text received:", fmt.Sprintf("%v", []byte(tmp)))
        }

        if result["mfont"] == "" {
                n.font = "Droid Sans Fallback"
                logger.Printf("mfont empty", n.font )
        } else {
                n.font, err = base64UrlDecode(strings.Split(result["mfont"], "/")[2])
                if err != nil {
                        return WaterMarkTask{}, errors.New("mfont decode error")
                }
        }
        
        if result["msize"] == "" {
                n.size = "200"
                logger.Printf("msize empty",n.size )
        } else {
                n.size = strings.Split(result["msize"], "/")[2]
        }

        if result["mo"] == "" {
                n.opacity = 1.0
                logger.Printf("opacity empty",n.opacity )
        } else {
                opacity,_ := strconv.ParseFloat(strings.Split(result["mo"], "/")[2],0)
                n.opacity = float32(opacity/100)
        }

        if result["mwidth"] == "" {
                n.width = 100
                logger.Printf("width empty",n.width )
        } else {
                n.width,_ = strconv.Atoi(strings.Split(result["mwidth"], "/")[2])
        }

        if result["mcolor"] == "" {
                n.color = bimg.Color{0,0,0}
                logger.Printf("color empty",n.color )
        } else {
                color := strings.Split(result["mcolor"], "/")[2]
                r, _ := strconv.ParseInt(color[:2], 16, 0)
                g, _ := strconv.ParseInt(color[2:4], 16, 0)
                b, _ := strconv.ParseInt(color[4:6], 16, 0)
                n.color = bimg.Color{uint8(r),uint8(g),uint8(b)}
        }

        if result["mx"] == "" {
                n.x = -10
                logger.Printf("nx empty",n.x )
        } else {
                n.x, _ = strconv.Atoi(strings.Split(result["mx"], "/")[2]) //convert to int
        }

        if result["my"] == "" {
                n.y = -10
                logger.Printf("ny empty",n.y )
        } else {
                n.y, _ = strconv.Atoi(strings.Split(result["my"], "/")[2]) //convert to int
        }
        return n, nil
}
func adjustCropTask(buffer *[]byte, plan *CropTask) () {
        img := bimg.NewImage(*buffer)
        s, _:= img.Size()
        xwidth := s.Width
        xheight := s.Height
        if plan.edge == 0  {//long edge policy
                if xwidth > xheight {
                        plan.height = 0
                } else if xwidth < xheight {
                        plan.width = 0
                }
        }

        if plan.edge == 1  {//short edge policy
                if xwidth > xheight {
                        plan.width = 0
                } else if xwidth < xheight {
                        plan.height = 0
                }
        }
}

func ProcessImage(filename string, plan *CropTask) ([]byte, error) {
        logger.Println("start to Process plan",plan)
        buffer, _ := bimg.Read(filename)
        img := bimg.NewImage(buffer)
        s, _:= img.Size()
        xwidth := s.Width
        xheight := s.Height
    var o bimg.Options
        var err error
        var new []byte 
        switch plan.mode {
        case 0:
                if (plan.large == 1) {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Enlarge:true}        
                } else {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Enlarge:false}        
                }
                new, err = bimg.Resize(buffer, o)
        case 1:
                adjustCropTask(&buffer, plan)
                logger.Println("now plan",plan)
                if (plan.large == 1) {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Enlarge:true}        
                } else {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Enlarge:false}        
                }
                new, err = bimg.Resize(buffer, o)
        case 2:
                if (plan.large == 1) {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Force:true}        
                } else {
                        if xwidth*xheight < plan.width*plan.height {
                                o = bimg.Options{Width:xwidth, Height:xheight, Force:true}
                        } else {
                                o = bimg.Options{Width:plan.width, Height:plan.height, Force:true}
                        }
                }
                new, err = bimg.Resize(buffer, o)
                        
        case 3:
                factor:= float64(plan.proportion)/100.0
                logger.Println("zoo factor :",factor)
                o = bimg.Options{Width:int(float64(xwidth)*factor), Height:int(float64(xheight)*factor), Force:true}
                new, err = bimg.Resize(buffer, o)
        case 4:
                o = bimg.Options{Width:plan.width, Height:plan.height, Embed:true, Enlarge:true}
                new, err = bimg.Resize(buffer,o)
        case 5:
                if (plan.large == 1) {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Crop: true, Enlarge:true}
                } else {
                        o = bimg.Options{Width:plan.width, Height:plan.height, Crop: true, Enlarge:false}
                }
                new, err = bimg.Resize(buffer,o)
        }
        if err != nil {
                fmt.Fprintln(os.Stderr, err)
        }
        return new, err
}

func ProcessWaterMark(filename string, object string, plan *WaterMarkTask) ([]byte, error) {
        logger.Println("start to Process plan",plan)
        buffer, _ := bimg.Read(filename)
        var err error
        var new []byte 
        switch plan.mode {
        case 0:
        case 1:
                logger.Println("now plan",plan)
                w := bimg.Watermark{
                        NoReplicate:true,
                        Text:plan.text,
                        Font:plan.font+" "+plan.size,
                        Opacity:plan.opacity,
                        Width:plan.width,
                        DPI:30,
                        Margin:0,
                        Background:plan.color,
                        Xoffset:plan.x,
                        Yoffset:plan.y,
                }
                logger.Println("now watermark",w)
                new, err = bimg.NewImage(buffer).Watermark(w)
        case 2:
        }
                
        if err != nil {
                fmt.Fprintln(os.Stderr, err)
        }
        return new, err
}
func download(client *http.Client, download_url string, uuid string) (string, int) {

                logger.Printf("start to download %s\n", download_url)
                resp, err := client.Get(download_url)
                if err != nil {
                        logger.Println("download failed")
                        return "", 404
                }
                defer resp.Body.Close()

                //check header
                if resp.StatusCode != 200 {
                        logger.Println("Request is not 200")
                        return "", resp.StatusCode
                }


                mime_type := resp.Header.Get("Content-Type")

                if strings.Contains(mime_type, "image") == false {
                        if ok, _ := regexp.MatchString("(jpeg|jpg|png|gif)", download_url) ; ok == false {
                                logger.Printf("MIME TYPE is %s not an image\n", mime_type)
                                return "", http.StatusUnsupportedMediaType //415
                        }
                }

                content_length := resp.Header.Get("Content-Length")

                if len, _ := strconv.Atoi(content_length); len > (20<<20) {
                        return "", http.StatusRequestEntityTooLarge
                }


                /* open temp file */
                tmpfile, err := ioutil.TempFile("",uuid)
                if err != nil {
                        logger.Printf("can not create temp file %s", uuid)
                        return "", 404
                }
                defer tmpfile.Close()

                n, err := io.Copy(tmpfile, resp.Body)
                if err != nil {
                        return "", 404
                }

                logger.Printf("download %d bytes from %s OK\n", n, download_url)
                return tmpfile.Name(), 200
}

var UNKNOWN string ="unknown"

func Slave(taskQ chan string, resultQ chan FinishTask, client * http.Client, slave_num int) {

        for {
                task := <-taskQ
                //split your url
                var data TaskData
                dec := json.NewDecoder(strings.NewReader(task))
                if err := dec.Decode(&data); err != nil {
                        logger.Println("Decode failed")
                        returnError(400, UNKNOWN,"", resultQ)
                        continue
                }
                uuid := data.Uuid
                url := data.Url
                logger.Printf("I got task %s %s\n", uuid, url)

                //download content from data
                //stripe all the query string and add "http://"
                //find the first ?
                taskType := -1
                var pos int

                if pos_0 := strings.Index(url, "@imageview/"); pos_0 != -1 {
                        taskType = IMAGEVIEW
                        pos = pos_0
                }

                if pos_1 := strings.Index(url, "@watermark/"); pos_1 != -1 {
                        taskType = WATERMARK
                        pos = pos_1
                }

                if taskType == -1 {
                        logger.Printf("can not found convert parameters")
                        returnError(400,uuid,url, resultQ)
                        continue
                }

                //if remove any slash at the start
                var start_pos int
                var v rune
                for start_pos, v = range url[0:pos] {
                        if string(v) !="/" {
                                break
                        }
                }

                download_url := "http://" + url[start_pos:pos]
                convert_params := url[pos+1:]
                // parse convert_params
                // now only suport 
                // ?imageView/1/w/<Width>/h/<Height>  width is at least, Height is at least, crop at center 
                // if only either is specified, use the same size;
                // ?imageView/2/w/<Width>/h/<Height>  width is at most,  Height is as most
                // ?imageView/3/w/<Width>/h/<Height>  width is as least, Height is as least
                
                var r []string
                var names []string
                switch taskType {
                        case IMAGEVIEW:
                                r = imageviewPattern.FindStringSubmatch(convert_params)
                                names = imageviewNames
                        case WATERMARK:
                                r = watermarkPattern.FindStringSubmatch(convert_params)
                                names = watermarkNames
                } 

                captures := make(map[string]string)
                for i, name := range names{
                        if i == 0 {
                                continue
                        }
                        logger.Println("name:%s,%s",name,r[i])
                        captures[name] = r[i]
                }
                
                if taskType == IMAGEVIEW {
                        plan, err := PreProcess(captures)

                        if err != nil {
                                returnError(400,uuid,url, resultQ)
                                continue
                        }

                        //seems convert command OK, plan a process
                        //

                        /* now start to download the image to a temperary directoy */
                        origin_filename, code := download(client, download_url, uuid)
                        if code != 200 {
                                returnError(code,uuid,url, resultQ)
                                os.Remove(origin_filename)
                                continue
                        }


                        processed_blob, err := ProcessImage(origin_filename, &plan)
                        os.Remove(origin_filename)
                        if err != nil {
                                returnError(400,uuid,url, resultQ)
                                continue
                        }
                        //logger.Println(processed_blob)

                        //write_to_local_file(uuid, processed_blob)
                        resultQ <- FinishTask{200,uuid,url, processed_blob}
                } else if taskType == WATERMARK {
                        plan, err := PreProcessWaterMark(captures)

                        if err != nil {
                                returnError(400,uuid,url, resultQ)
                                continue
                        }

                        //seems convert command OK, plan a process
                        //

                        /* now start to download the image to a temperary directoy */
                        origin_filename, code := download(client, download_url, uuid)
                        if code != 200 {
                                returnError(code,uuid,url, resultQ)
                                os.Remove(origin_filename)
                                continue
                        }
                        
                        var object_filename string
                        /* now start to download the object to a temperary directoy */
                        if plan.mode != 1 {
                                object_filename, code := download(client, plan.object, uuid+".object")
                                if code != 200 {
                                        returnError(code,uuid,url, resultQ)
                                        os.Remove(origin_filename)
                                        os.Remove(object_filename)
                                        continue
                                }
                        }

                        processed_blob, err := ProcessWaterMark(origin_filename, object_filename, &plan)
                        os.Remove(origin_filename)
                        if plan.mode != 1 {
                                os.Remove(object_filename)
                        }
                        if err != nil {
                                returnError(400,uuid,url, resultQ)
                                continue
                        }
                        //logger.Println(processed_blob)

                        //write_to_local_file(uuid, processed_blob)
                        resultQ <- FinishTask{200,uuid,url, processed_blob}

                }

        }
}


func write_to_local_file(uuid string, blob []byte) {
                tmpfile, _ := ioutil.TempFile("", uuid)
                tmpfile.Write(blob)
                tmpfile.Close()
}



func reportFinish(resultQ chan FinishTask, pool *redis.Pool) {
        redis_conn := pool.Get()
        defer redis_conn.Close()
        for r := range resultQ {
                //put data back to redis
                if r.code == 200 {
                        redis_conn.Do("MULTI")
                        redis_conn.Do("SET", r.url, r.blob)
                        redis_conn.Do("LPUSH", r.uuid, r.code)
                        redis_conn.Do("EXEC")
                        r.blob = nil
                } else {
                        redis_conn.Do("LPUSH", r.uuid, r.code)
                }
                logger.Printf("finishing task [%s] for %s code %d\n", r.uuid, r.url, r.code)
        }
}

/*https://godoc.org/github.com/garyburd/redigo/redis#Pool*/
func newRedisPool(server, password string) *redis.Pool {
    return &redis.Pool{
        MaxIdle: 3,
        IdleTimeout: 60 * time.Second,
        Dial: func () (redis.Conn, error) {
            c, err := redis.Dial("tcp", server)
            if err != nil {
                return nil, err
            }
            return c, err
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
    }
}


func main() {


        go func() {
                        log.Println(http.ListenAndServe("localhost:6060", nil))
        }()


        
            app := cli.NewApp()
        app.Name = "crop image daemon"
        app.Usage = "--redis-server --redis-port"
        app.Flags = []cli.Flag{
                cli.StringFlag{
                Name:   "redis-server",
                Usage:  "redis server address",
                EnvVar: "REDIS_SERVER",
                Value:  "127.0.0.1",
                }, 
                cli.StringFlag{
                Name:   "redis-port",
                Usage:  "redis port address",
                EnvVar: "REDIS_PORT",
                Value:  "6379",
                }, 
        }        
	
        var redis_address string;
        app.Action = func(c *cli.Context) error {
            redis_address = c.String("redis-server") + ":" + c.String("redis-port")
	    return nil 
        }

        app.Run(os.Args)

        var pool *redis.Pool
        pool = newRedisPool(redis_address, "")

        redis_con := pool.Get()
        defer redis_con.Close()

        file, err := os.Create("/var/log/cropimage.log")
        if err != nil {
                fmt.Println("failed open log")
                return
        }
        //Redirect stdout and stderr to the log
        syscall.Dup2(int(file.Fd()), 2)
        syscall.Dup2(int(file.Fd()), 1)

        defer file.Close()
        logger = log.New(file, "logger: ", log.LstdFlags)

        taskQ := make(chan string, 10)
        returnQ := make(chan FinishTask)

        //generate a pool of works
        //create http client pools
        httpClient := &http.Client{Timeout: time.Second * 5}

        numOfWorkers := 50

        for i := 0; i< numOfWorkers; i++ {
                go Slave(taskQ, returnQ, httpClient, i)
        }
        go reportFinish(returnQ, pool)

        //will use signal channel to quit
        for {
                r, err := redis.Strings(redis_con.Do("BLPOP","taskQueue", 0))
                if err != nil {
                        logger.Printf("something bad happend %v",err)
                        return
                }
                logger.Println("Now have", r[1])
                taskQ <- r[1]
        }
}
