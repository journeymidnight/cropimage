package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/thesues/bimg"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type TaskData struct {
	Uuid string `json:"uuid"`
	Url  string `json:"url"`
}

type ReturnData struct {
	Blob []byte `json:"blob"`
	Mime string `json:"mime"`
}

type CropTask struct {
	mode       string
	width      int
	height     int
	long       int
	short      int
	limit      int
	color      bimg.Color
	proportion int
}

type WaterMarkTask struct {
	mode     int
	object   string
	text     string
	font     string
	size     string
	color    bimg.Color
	order    int
	align    int
	interval int
	position int
	opacity  float32
	width    int
	x        int
	y        int
}

type FinishTask struct {
	code int
	uuid string
	url  string
	blob []byte
	mime string
}

const (
	RESIZE    = "resize"
	WATERMARK = "watermark"
)

var logger *log.Logger
var resizePattern = regexp.MustCompile("resize(?P<m>,m_[a-z]+)?(?P<w>,w_[0-9]+)?(?P<h>,h_[0-9]+)?(?P<l>,l_[0-9]+)?(?P<s>,s_[0-9]+)?(?P<limit>,limit_[0-1])?(?P<color>,color_[0-9A-F]+)?(?P<p>,p_[0-9]+)?")
var watermarkPattern = regexp.MustCompile("watermark(?P<t>,t_[0-9]+)?(?P<g>,g_[a-z]+)?(?P<x>,x_[0-9]+)?(?P<y>,y_[0-9]+)?(?P<voffset>,voffset_[-0-9]+)?(?P<text>,text_[a-zA-Z0-9-_=]+)?(?P<type>,type_[a-zA-Z0-9-_=]+)?(?P<color>,color_[0-9A-F]{6})?(?P<size>,size_[0-9]+)?(?P<shadow>,shadow_[0-9]+)?(?P<rotate>,rotate_[0-9]+)?(?P<fill>,fill_[0-1])?")
var resizeNames = resizePattern.SubexpNames()
var watermarkNames = watermarkPattern.SubexpNames()

func returnError(code int, uuid string, url string, Q chan FinishTask) {
	Q <- FinishTask{code, uuid, url, nil, ""}
}

func returnUnchange(code int, uuid string, url string, filePath string, Q chan FinishTask) {
	buffer, _ := bimg.Read(filePath)
	img := bimg.NewImage(buffer)
	Q <- FinishTask{code, uuid, url, buffer, img.Type()}
}

func PreProcess(result map[string]string) (CropTask, error) {

	n := CropTask{}

	if result["p"] == "" {
		n.proportion = 0
	} else {
		n.proportion, _ = strconv.Atoi(result["p"])
		if n.proportion < 1 || n.proportion > 1000 {
			return CropTask{}, errors.New("wrong resize p detect")
		}
		return n, nil
	}

	if result["w"] == "" {
		n.width = 0
	} else {
		n.width, _ = strconv.Atoi(result["w"])
		if n.width < 1 || n.width > 4096 {
			return CropTask{}, errors.New("wrong resize width detect")
		}
	}

	if result["h"] == "" {
		n.height = 0
	} else {
		n.height, _ = strconv.Atoi(result["h"])
		if n.height < 1 || n.height > 4096 {
			return CropTask{}, errors.New("wrong resize height detect")
		}
	}

	if result["l"] == "" {
		n.long = 0
	} else {
		n.long, _ = strconv.Atoi(result["l"])
		if n.long < 1 || n.long > 4096 {
			return CropTask{}, errors.New("wrong resize long detect")
		}
	}

	if result["s"] == "" {
		n.short = 0
	} else {
		n.short, _ = strconv.Atoi(result["s"])
		if n.short < 1 || n.short > 4096 {
			return CropTask{}, errors.New("wrong resize short detect")
		}
	}

	if result["limit"] == "" {
		n.limit = 1
	} else {
		n.limit, _ = strconv.Atoi(result["limit"])
		if n.limit != 0 && n.limit != 1 {
			return CropTask{}, errors.New("wrong resize limit detect")
		}
	}

	if result["color"] == "" {
		n.color = bimg.Color{255, 255, 255}
	} else {
		r, _ := strconv.ParseInt(result["color"][:2], 16, 0)
		g, _ := strconv.ParseInt(result["color"][2:4], 16, 0)
		b, _ := strconv.ParseInt(result["color"][4:6], 16, 0)
		n.color = bimg.Color{uint8(r), uint8(g), uint8(b)}
	}

	switch result["m"] {
	case "", "lfit":
		n.mode = "lfit"
		if ((n.width != 0 || n.height != 0) && (n.long != 0 || n.short != 0)) == true {
			return CropTask{}, errors.New("can not resize in height&width and long&short at the same time")
		}
	case "mfit":
		n.mode = "mfit"
		if ((n.width != 0 || n.height != 0) && (n.long != 0 || n.short != 0)) == true {
			return CropTask{}, errors.New("can not resize in height&width and long&short at the same time")
		}
	case "fill", "pad", "fixed":
		n.mode = result["m"]
		if n.width != 0 && n.height == 0 {
			n.height = n.width
		}

		if n.width == 0 && n.height != 0 {
			n.width = n.height
		}
	default:
		return CropTask{}, errors.New("wrong resize mode detect")
	}
	//if n.mode == "lfit" || n.mode == "mfit" {
	//	if n.width == 0 && n.height == 0 {
	//		return CropTask{}, errors.New("width and hight can not be both empty in mode lfit or mfit")
	//	}
	//}
	//
	//if n.mode == "lfit" || n.mode == "mfit" {
	//	if n.width == 0 && n.height == 0 {
	//		return CropTask{}, errors.New("width and hight can not be both empty in mode lfit or mfit")
	//	}
	//}
	//
	//if n.mode == 1 || n.mode == 2 {
	//	if n.width == 0 || n.height == 0 {
	//		return CropTask{}, errors.New("must specify width and hight both in mode 1 and mode2")
	//	}
	//}
	//
	//if n.mode == 3 {
	//	if n.proportion < 1 || n.proportion > 1000 {
	//		return CropTask{}, errors.New("proportion exceed limitation")
	//	}
	//}
	return n, nil
}

func base64UrlDecode(src string) (string, error) {
	str := strings.Replace(src, "-", "+", -1)
	str = strings.Replace(str, "_", "/", -1)
	tmp, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(tmp), nil
}

//func PreProcessWaterMark(result map[string]string) (WaterMarkTask, error) {
//
//	n := WaterMarkTask{}
//	var tmp []byte
//	var err error
//	if result["mode"] == "" {
//		logger.Printf("no watermark mode")
//		return WaterMarkTask{}, errors.New("no watermark data")
//	}
//
//	n.mode, _ = strconv.Atoi(result["mode"]) //convert to int
//
//	if result["mobject"] == "" {
//		n.object = ""
//		logger.Printf("mobject empty", n.object)
//	} else {
//		n.object, err = base64UrlDecode(strings.Split(result["mobject"], "/")[2])
//		if err != nil {
//			return WaterMarkTask{}, errors.New("mobject decode error")
//		}
//	}
//
//	if result["mtext"] == "" {
//		logger.Printf("mtext empty", n.text)
//		n.text = ""
//	} else {
//		n.text, err = base64UrlDecode(strings.Split(result["mtext"], "/")[2])
//		if err != nil {
//			return WaterMarkTask{}, errors.New("mtext decode error")
//		}
//		logger.Printf("text received:", fmt.Sprintf("%v", []byte(tmp)))
//	}
//
//	if result["mfont"] == "" {
//		n.font = "Droid Sans Fallback"
//		logger.Printf("mfont empty", n.font)
//	} else {
//		n.font, err = base64UrlDecode(strings.Split(result["mfont"], "/")[2])
//		if err != nil {
//			return WaterMarkTask{}, errors.New("mfont decode error")
//		}
//	}
//
//	if result["msize"] == "" {
//		n.size = "200"
//		logger.Printf("msize empty", n.size)
//	} else {
//		n.size = strings.Split(result["msize"], "/")[2]
//	}
//
//	if result["mo"] == "" {
//		n.opacity = 1.0
//		logger.Printf("opacity empty", n.opacity)
//	} else {
//		opacity, _ := strconv.ParseFloat(strings.Split(result["mo"], "/")[2], 0)
//		n.opacity = float32(opacity / 100)
//	}
//
//	if result["mwidth"] == "" {
//		n.width = 100
//		logger.Printf("width empty", n.width)
//	} else {
//		n.width, _ = strconv.Atoi(strings.Split(result["mwidth"], "/")[2])
//	}
//
//	if result["mcolor"] == "" {
//		n.color = bimg.Color{0, 0, 0}
//		logger.Printf("color empty", n.color)
//	} else {
//		color := strings.Split(result["mcolor"], "/")[2]
//		r, _ := strconv.ParseInt(color[:2], 16, 0)
//		g, _ := strconv.ParseInt(color[2:4], 16, 0)
//		b, _ := strconv.ParseInt(color[4:6], 16, 0)
//		n.color = bimg.Color{uint8(r), uint8(g), uint8(b)}
//	}
//
//	if result["mx"] == "" {
//		n.x = -10
//		logger.Printf("nx empty", n.x)
//	} else {
//		n.x, _ = strconv.Atoi(strings.Split(result["mx"], "/")[2]) //convert to int
//	}
//
//	if result["my"] == "" {
//		n.y = -10
//		logger.Printf("ny empty", n.y)
//	} else {
//		n.y, _ = strconv.Atoi(strings.Split(result["my"], "/")[2]) //convert to int
//	}
//	return n, nil
//}

func adjustCropTask(buffer *[]byte, plan *CropTask) {
	img := bimg.NewImage(*buffer)
	s, _ := img.Size()
	xwidth := s.Width
	xheight := s.Height

	//单宽高缩放
	if plan.width+plan.height != 0 && plan.width*plan.height == 0 {
		return
	}
	//单长短边缩放
	if plan.long+plan.short != 0 && plan.long*plan.short == 0 {
		if plan.long != 0 {
			if xwidth >= xheight {
				plan.width = plan.long
				plan.height = 0
			} else {
				plan.height = plan.long
				plan.width = 0
			}
		} else {
			if xwidth >= xheight {
				plan.height = plan.short
				plan.width = 0
			} else {
				plan.width = plan.short
				plan.height = 0
			}
		}
		return
	}

	//同时指定宽高缩放
	if plan.width > 0 && plan.height > 0 {
		if plan.mode == "lfit" { //长边优先
			if xwidth >= xheight {
				plan.height = 0
			} else {
				plan.width = 0
			}
		}

		if plan.mode == "mfit" { //短边优先
			if xwidth >= xheight {
				plan.width = 0
			} else {
				plan.height = 0
			}
		}
		return
	}

	//同时指定长短边缩放
	if plan.long > 0 && plan.short > 0 {
		if plan.mode == "lfit" { //长边优先
			if xwidth >= xheight {
				plan.width = plan.long
				plan.height = 0
			} else {
				plan.height = plan.long
				plan.width = 0
			}
		}

		if plan.mode == "mfit" { //短边优先
			if xwidth >= xheight {
				plan.height = plan.short
				plan.width = 0
			} else {
				plan.width = plan.short
				plan.height = 0
			}
		}
		return
	}
	return
}

func ProcessImage(filename string, plan *CropTask) ([]byte, string, error) {
	logger.Println("start to Process plan", plan)
	buffer, _ := bimg.Read(filename)
	img := bimg.NewImage(buffer)
	s, _ := img.Size()
	mime := img.Type()
	xwidth := s.Width
	xheight := s.Height
	var o bimg.Options
	var err error
	var new []byte

	//比例缩放
	if plan.proportion != 0 {
		factor := float64(plan.proportion) / 100.0
		logger.Println("zoo factor :", factor)
		o = bimg.Options{Width: int(float64(xwidth) * factor), Height: int(float64(xheight) * factor), Force: true}
		new, err = bimg.Resize(buffer, o)
		return new, mime, err
	}

	switch plan.mode {
	//长边优先
	case "lfit":
		adjustCropTask(&buffer, plan)
		if plan.limit == 0 {
			o = bimg.Options{Width: plan.width, Height: plan.height, Enlarge: true}
		} else {
			o = bimg.Options{Width: plan.width, Height: plan.height, Enlarge: false}
		}
		new, err = bimg.Resize(buffer, o)
	//短边优先
	case "mfit":
		adjustCropTask(&buffer, plan)
		logger.Println("now plan", plan)
		if plan.limit == 0 {
			o = bimg.Options{Width: plan.width, Height: plan.height, Enlarge: true}
		} else {
			o = bimg.Options{Width: plan.width, Height: plan.height, Enlarge: false}
		}
		new, err = bimg.Resize(buffer, o)
	//case "fill":
	//	if plan.large == 1 {
	//		o = bimg.Options{Width: plan.width, Height: plan.height, Force: true}
	//	} else {
	//		if xwidth*xheight < plan.width*plan.height {
	//			o = bimg.Options{Width: xwidth, Height: xheight, Force: true}
	//		} else {
	//			o = bimg.Options{Width: plan.width, Height: plan.height, Force: true}
	//		}
	//	}
	//	new, err = bimg.Resize(buffer, o)
	case "pad":
		if plan.limit == 0 {
			o = bimg.Options{Width: plan.width, Height: plan.height, Background: plan.color, Embed: true, Enlarge: true}
		} else {
			o = bimg.Options{Width: plan.width, Height: plan.height, Background: plan.color, Embed: true, Enlarge: false}
		}
		new, err = bimg.Resize(buffer, o)
	case "fixed":
		if plan.limit == 0 {
			o = bimg.Options{Width: plan.width, Height: plan.height, Force: true, Enlarge: true}
		} else {
			o = bimg.Options{Width: plan.width, Height: plan.height, Force: true, Enlarge: false}
		}
		new, err = bimg.Resize(buffer, o)
	case "fill":
		if plan.limit == 0 {
			o = bimg.Options{Width: plan.width, Height: plan.height, Crop: true, Enlarge: true}
		} else {
			o = bimg.Options{Width: plan.width, Height: plan.height, Crop: true, Enlarge: false}
		}
		new, err = bimg.Resize(buffer, o)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return new, mime, err
}

//func ProcessWaterMark(filename string, object string, plan *WaterMarkTask) ([]byte, error) {
//	logger.Println("start to Process plan", plan)
//	buffer, _ := bimg.Read(filename)
//	var err error
//	var new []byte
//	switch plan.mode {
//	case 0:
//	case 1:
//		logger.Println("now plan", plan)
//		w := bimg.Watermark{
//			NoReplicate: true,
//			Text:        plan.text,
//			Font:        plan.font + " " + plan.size,
//			Opacity:     plan.opacity,
//			Width:       plan.width,
//			DPI:         30,
//			Margin:      0,
//			Background:  plan.color,
//			Xoffset:     plan.x,
//			Yoffset:     plan.y,
//		}
//		logger.Println("now watermark", w)
//		new, err = bimg.NewImage(buffer).Watermark(w)
//	case 2:
//	}
//
//	if err != nil {
//		fmt.Fprintln(os.Stderr, err)
//	}
//	return new, err
//}

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
		if ok, _ := regexp.MatchString("(jpeg|jpg|png|gif)", download_url); ok == false {
			logger.Printf("MIME TYPE is %s not an image\n", mime_type)
			return "", http.StatusUnsupportedMediaType //415
		}
	}

	content_length := resp.Header.Get("Content-Length")

	if len, _ := strconv.Atoi(content_length); len > (20 << 20) {
		return "", http.StatusRequestEntityTooLarge
	}

	/* open temp file */
	tmpfile, err := ioutil.TempFile("", uuid)
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

var UNKNOWN string = "unknown"

func Slave(taskQ chan string, resultQ chan FinishTask, client *http.Client, slave_num int) {

	for {
		task := <-taskQ
		//split your url
		var data TaskData
		dec := json.NewDecoder(strings.NewReader(task))
		if err := dec.Decode(&data); err != nil {
			logger.Println("Decode failed")
			returnError(400, UNKNOWN, "", resultQ)
			continue
		}
		uuid := data.Uuid
		url := data.Url
		logger.Printf("I got task %s %s\n", uuid, url)

		//download content from data
		//stripe all the query string and add "http://"
		//find the first ?
		taskType := ""
		var pos int
		var pType int

		if pos_0 := strings.Index(url, "?x-oss-process=image/"); pos_0 != -1 {
			pos = pos_0
			pType = 0
		} else if pos_1 := strings.Index(url, "?x-oss-process=style/"); pos_1 != -1 {
			pos = pos_1
			pType = 1
		} else {
			logger.Printf("can not found convert parameters")
			returnError(400, uuid, url, resultQ)
			continue
		}

		//if taskType == -1 {
		//        logger.Printf("can not found convert parameters")
		//        returnError(400,uuid,url, resultQ)
		//        continue
		//}

		//if remove any slash at the start
		var start_pos int
		var v rune
		for start_pos, v = range url[0:pos] {
			if string(v) != "/" {
				break
			}
		}

		download_url := "http://" + url[start_pos:pos]
		convert_params := url[pos+len("?x-oss-process=image/"):]
		convert_params_slice := strings.Split(convert_params, "/")

		origin_filename, retCode := download(client, download_url, uuid)
		if retCode != 200 {
			returnError(retCode, uuid, url, resultQ)
			os.Remove(origin_filename)
			continue
		}

		if pType == 1 {
			returnUnchange(200, uuid, url, origin_filename, resultQ)
			os.Remove(origin_filename)
			continue
		}

		for _, task := range convert_params_slice {
			var r []string
			var names []string

			if strings.HasPrefix(task, RESIZE) {
				r = resizePattern.FindStringSubmatch(convert_params)
				names = resizeNames
				taskType = RESIZE
			} else if strings.HasPrefix(task, WATERMARK) {
				r = watermarkPattern.FindStringSubmatch(convert_params)
				names = watermarkNames
				taskType = WATERMARK
			} else {
				continue
			}

			captures := make(map[string]string)
			for i, name := range names {
				if i == 0 {
					continue
				}
				logger.Println("name:%s,%s", name, r[i])
				splited := strings.Split(r[i], "_")
				if len(splited) < 2 {
					captures[name] = ""
				} else {
					captures[name] = splited[1]
				}
			}

			if taskType == RESIZE {
				plan, err := PreProcess(captures)

				if err != nil {
					returnUnchange(200, uuid, url, origin_filename, resultQ)
					os.Remove(origin_filename)
					continue
				}

				//seems convert command OK, plan a process
				//

				///* now start to download the image to a temperary directoy */
				//if origin_filename == "" {
				//	origin_filename, retCode = download(client, download_url, uuid)
				//	if retCode != 200 {
				//		returnError(retCode, uuid, url, resultQ)
				//		os.Remove(origin_filename)
				//		continue
				//	}
				//}

				processed_blob, mime, err := ProcessImage(origin_filename, &plan)
				if err != nil {
					returnUnchange(200, uuid, url, origin_filename, resultQ)
					os.Remove(origin_filename)
					continue
				}
				//logger.Println(processed_blob)
				os.Remove(origin_filename)

				//write_to_local_file(uuid, processed_blob)
				resultQ <- FinishTask{200, uuid, url, processed_blob, mime}
			} else if taskType == WATERMARK {
				//plan, err := PreProcessWaterMark(captures)
				//
				//if err != nil {
				//	returnError(400, uuid, url, resultQ)
				//	continue
				//}
				////seems convert command OK, plan a process
				////
				///* now start to download the image to a temperary directoy */
				////if origin_filename == "" {
				////	origin_filename, retCode = download(client, download_url, uuid)
				////	if retCode != 200 {
				////		returnError(retCode, uuid, url, resultQ)
				////		os.Remove(origin_filename)
				////		continue
				////	}
				////}
				//var object_filename string
				///* now start to download the object to a temperary directoy */
				//if plan.mode != 1 {
				//	if object_filename == "" {
				//		object_filename, code := download(client, plan.object, uuid+".object")
				//		if code != 200 {
				//			returnError(code, uuid, url, resultQ)
				//			os.Remove(origin_filename)
				//			os.Remove(object_filename)
				//			continue
				//		}
				//	}
				//}
				//
				//processed_blob, err := ProcessWaterMark(origin_filename, object_filename, &plan)
				//os.Remove(origin_filename)
				//if plan.mode != 1 {
				//	os.Remove(object_filename)
				//}
				//if err != nil {
				//	returnError(400, uuid, url, resultQ)
				//	continue
				//}
				//logger.Println(processed_blob)

				//write_to_local_file(uuid, processed_blob)
				//resultQ <- FinishTask{200, uuid, url, processed_blob}
				returnUnchange(200, uuid, url, origin_filename, resultQ)
				os.Remove(origin_filename)
				continue
			}
		}
		// parse convert_params
		// now only suport
		// ?imageView/1/w/<Width>/h/<Height>  width is at least, Height is at least, crop at center
		// if only either is specified, use the same size;
		// ?imageView/2/w/<Width>/h/<Height>  width is at most,  Height is as most
		// ?imageView/3/w/<Width>/h/<Height>  width is as least, Height is as least

	}
}

func write_to_local_file(uuid string, blob []byte) {
	tmpfile, _ := ioutil.TempFile("", uuid)
	tmpfile.Write(blob)
	tmpfile.Close()
}

func combineData(blob []byte, mime string) []byte {
	var a [20]byte
	copy(a[:], mime)
	return append(a[:], blob[:]...)
}

func reportFinish(resultQ chan FinishTask, pool *redis.Pool) {
	redis_conn := pool.Get()
	defer redis_conn.Close()
	for r := range resultQ {
		//put data back to redis
		if r.code == 200 {
			//rd := ReturnData{r.blob, r.mime}
			//combine, _ := json.Marshal(&rd)
			combined := combineData(r.blob, r.mime)
			redis_conn.Do("MULTI")
			redis_conn.Do("SET", r.url, combined)
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
		MaxIdle:     3,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
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

	var redis_address string
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

	for i := 0; i < numOfWorkers; i++ {
		go Slave(taskQ, returnQ, httpClient, i)
	}
	go reportFinish(returnQ, pool)

	//will use signal channel to quit
	for {
		r, err := redis.Strings(redis_con.Do("BLPOP", "taskQueue", 0))
		if err != nil {
			logger.Printf("something bad happend %v", err)
			return
		}
		logger.Println("Now have", r[1])
		taskQ <- r[1]
	}
}
