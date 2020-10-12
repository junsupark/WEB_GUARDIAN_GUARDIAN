package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strings"
	"time"

	gosocketio "github.com/graarh/golang-socketio"
	"gocv.io/x/gocv"
)

var ViewChannel = make(chan []byte)
var encodingSize image.Point = image.Point{854, 480}

type DetectPointInfo struct {
	ViewSize    image.Point
	DetectPoint [][]image.Point
}

// getOutputsNames : YOLO Layer
func getOutputsNames(net *gocv.Net) []string {
	var outputLayers []string
	for _, i := range net.GetUnconnectedOutLayers() {
		layer := net.GetLayer(i)
		layerName := layer.GetName()
		if layerName != "_input" {
			outputLayers = append(outputLayers, layerName)
		}
	}
	return outputLayers
}

// PostProcess : All Detect Box
func PostProcess(frame gocv.Mat, outs *[]gocv.Mat) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle
	for _, out := range *outs {

		data, _ := out.DataPtrFloat32()
		for i := 0; i < out.Rows(); i, data = i+1, data[out.Cols():] {

			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(5, out.Cols())
			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)
			if confidence > 0.5 {

				centerX := int(data[0] * float32(frame.Cols()))
				centerY := int(data[1] * float32(frame.Rows()))
				width := int(data[2] * float32(frame.Cols()))
				height := int(data[3] * float32(frame.Rows()))

				left := centerX - width/2
				top := centerY - height/2
				classIds = append(classIds, classIDPoint.X)
				confidences = append(confidences, float32(confidence))
				boxes = append(boxes, image.Rect(left, top, width, height))
			}
		}
	}
	return boxes, confidences, classIds
}

// ReadCOCO : Read coco.names
func ReadCOCO() []string {
	var classes []string
	read, _ := os.Open("./assets/coco.names")
	defer read.Close()
	for {
		var t string
		_, err := fmt.Fscan(read, &t)
		if err != nil {
			break
		}
		classes = append(classes, t)
	}
	return classes
}

// drawRect : Detect Class to Draw Rect
func drawRect(img gocv.Mat, boxes []image.Rectangle, classes []string, classIds []int, indices []int) (gocv.Mat, []string) {
	var detectClass []string
	for _, idx := range indices {
		if idx == 0 {
			continue
		}
		gocv.Rectangle(&img, image.Rect(boxes[idx].Max.X, boxes[idx].Max.Y, boxes[idx].Max.X+boxes[idx].Min.X, boxes[idx].Max.Y+boxes[idx].Min.Y), color.RGBA{255, 0, 0, 0}, 2)
		gocv.PutText(&img, classes[classIds[idx]], image.Point{boxes[idx].Max.X, boxes[idx].Max.Y + 30}, gocv.FontHersheyPlain, 5, color.RGBA{0, 0, 255, 0}, 3)
		detectClass = append(detectClass, classes[classIds[idx]])
	}
	return img, detectClass
}

// TransPos : Frontend ViewSize => CV Mat ViewSize
func TransPos(FrontInfo DetectPointInfo, CameraIdx int, CvViewSize image.Point) [][]image.Point {
	TransPoint := make([][]image.Point, 1)
	for x := 0; x < len(FrontInfo.DetectPoint[CameraIdx]); x++ {
		transX := FrontInfo.DetectPoint[CameraIdx][x].X * CvViewSize.X / FrontInfo.ViewSize.X
		transY := FrontInfo.DetectPoint[CameraIdx][x].Y * CvViewSize.Y / FrontInfo.ViewSize.Y
		TransPoint[0] = append(TransPoint[0], image.Pt(transX, transY))
	}
	return TransPoint
}

// Detect : Run YOLOv4 Process
func Detect(net *gocv.Net, src gocv.Mat, scoreThreshold float32, nmsThreshold float32, OutputNames []string, classes []string) (gocv.Mat, []string) {
	img := src.Clone()
	img.ConvertTo(&img, gocv.MatTypeCV32F)
	blob := gocv.BlobFromImage(img, 1/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
	net.SetInput(blob, "")
	probs := net.ForwardLayers(OutputNames)
	boxes, confidences, classIds := PostProcess(img, &probs)

	indices := make([]int, 100)
	if len(boxes) == 0 { // No Classes
		return src, []string{}
	}
	gocv.NMSBoxes(boxes, confidences, scoreThreshold, nmsThreshold, indices)

	return drawRect(img, boxes, classes, classIds, indices)
}

func MotionDetect(src gocv.Mat, mog2 gocv.BackgroundSubtractorMOG2) (gocv.Mat, int) {
	img := src.Clone()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	// first phase of cleaning up image, obtain foreground only
	mog2.Apply(img, &imgDelta)

	// remaining cleanup of the image to use for finding contours.
	// first use threshold
	gocv.Threshold(imgDelta, &imgDelta, 25, 255, gocv.ThresholdBinary)

	// then dilate
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(imgDelta, &imgDelta, kernel)

	// now find contours
	contours := gocv.FindContours(imgDelta, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	contours_cnt := 0
	for i, c := range contours {
		area := gocv.ContourArea(c)
		if area < 500 && area < 5000 {
			continue
		}

		gocv.DrawContours(&img, contours, i, color.RGBA{255, 0, 255, 0}, 2)

		rect := gocv.BoundingRect(c)
		gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
		contours_cnt++
	}
	if contours_cnt > 10 {
		contours_cnt = -1
	}
	return img, contours_cnt
}

func DetectArea(img gocv.Mat, info DetectPointInfo) gocv.Mat {

	imgClone := img.Clone()
	if len(info.DetectPoint) == 0 {
		return imgClone
	}
	mask := gocv.NewMatWithSize(imgClone.Rows(), imgClone.Cols(), gocv.MatTypeCV8UC1)

	result := gocv.NewMatWithSize(imgClone.Rows(), imgClone.Cols(), gocv.MatTypeCV8UC1)
	gocv.FillPoly(&mask, info.DetectPoint, color.RGBA{255, 255, 255, 0})
	imgClone.CopyToWithMask(&result, mask)
	return result
	boundingRect := gocv.BoundingRect(gocv.FindContours(mask.Clone(), gocv.RetrievalExternal, gocv.ChainApproxSimple)[0])
	return result.Region(boundingRect)
}

func DetectStart(CapUrl string, Server *gosocketio.Server, DetectPointChannel chan DetectPointInfo) {
	cap, err := gocv.OpenVideoCapture(CapUrl)
	if err != nil {
		fmt.Printf("Error opening capture device")
		return
	}
	defer cap.Close()

	// Motion Init
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	// YOLO Init
	net := gocv.ReadNet("./assets/yolov4.weights", "./assets/yolov4.cfg")
	defer net.Close()
	net.SetPreferableBackend(gocv.NetBackendType(gocv.NetBackendDefault))
	net.SetPreferableTarget(gocv.NetTargetType(gocv.NetTargetCPU))
	classes := ReadCOCO()
	OutputNames := getOutputsNames(&net)

	img := gocv.NewMat()
	defer img.Close()

	oriImg := gocv.NewMat()
	defer oriImg.Close()
	FrameChannel := make(chan gocv.Mat)
	YoloChannel := make(chan gocv.Mat)
	type IDetect struct {
		Thumbnail []byte `json:"thumbnail"`
		Content   string `json:"content"`
		Time      string `json:"time"`
	}

	go func() { // Yolo Thread
		for {
			q_img := <-YoloChannel
			detectImg, detectClass := Detect(&net, q_img, 0.45, 0.5, OutputNames, classes)
			buf, _ := gocv.IMEncode(".jpg", detectImg)
			fmt.Printf("class : %v\n ", detectClass)
			if len(detectClass) > 0 {
				b, _ := json.Marshal(IDetect{buf, strings.Join(detectClass, ","), time.Now().Format("2006-01-02 15:04:05")})
				Server.BroadcastToAll("detect", string(b))
			}
		}
	}()
	go func() { // Motion Detect Thread
		var startTime time.Time
		var endTime time.Duration
		var startFlag bool
		for img := range FrameChannel {
			//gocv.Resize(img, &img, image.Point{1280, 720}, 0, 0, 1)
			_, motionCnt := MotionDetect(img, mog2)
			if motionCnt > 0 { // 움직임 감지됐으면
				if !startFlag { // 움직임 감지 시작 시간 대입
					startFlag = true
					fmt.Println("감지 시작")
					startTime = time.Now()

					// go func() {
					// 	YoloChannel <- img
					// }()
				} else {
					ingTime := time.Since(startTime)
					var i time.Duration = 1
					for ; i < 30; i++ {
						if ingTime > time.Second*i && ingTime < time.Second*i+(time.Millisecond*10) { // 움직임 감지되고 1초 뒤
							// Process
							fmt.Printf("감지 %d초 됨.\n", i)

						}
					}
				}

			} else { // 움직임 감지없으면

				endTime = time.Since(startTime)

				if startFlag {

					if endTime > time.Second*3 { // 움직임 분기 (움직임 감지 3초 유지)
						// Process
						fmt.Printf("감지끝 %d초 이상 유지됨.\n", 3)
					}
					startFlag = false
				}

			}
			// buf, _ := gocv.IMEncode(".jpg", motion)
			// ViewChannel <- buf
		}
	}()
	var DPI DetectPointInfo
	go func() { // Set DetectPointInfo
		for D := range DetectPointChannel {
			DPI = DetectPointInfo{D.ViewSize, TransPos(D, 0, encodingSize)}
			fmt.Printf("DPI %v\n", DPI)
		}
	}()
	for {
		if ok := cap.Read(&img); !ok {
			log.Println("RTSP Close")
			cap, _ = gocv.OpenVideoCapture(CapUrl)
			log.Println("RTSP ReStart")
		}
		if img.Empty() {
			continue
		}

		gocv.Resize(img, &img, encodingSize, 0, 0, 0)
		buf, _ := gocv.IMEncode(".jpg", img)
		ViewChannel <- buf

		go func() {
			if DPI.ViewSize.X > 0 && DPI.ViewSize.Y > 0 { // 좌표 설정 돼있을 경우 이거 지우면 프로그램 안멈춤.
				roi := DetectArea(img, DPI)
				FrameChannel <- roi
			}
		}()
		gocv.WaitKey(1)
	}
}
