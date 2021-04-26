package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gocv.io/x/gocv"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func largestContour(contours gocv.PointsVector) int {
	largestIndex := -1
	largest := 0.0
	for i := 0; i < contours.Size(); i++ {
		area := gocv.ContourArea(contours.At(i))
		if area > largest {
			largest = area
			largestIndex = i
		}
	}
	return largestIndex
}

func imPoint2Pts(hull gocv.Mat, contour gocv.PointVector) []image.Point {
	hullPoints := []image.Point{}
	for i := 0; i < hull.Cols(); i++ {
		for j := 0; j < hull.Rows(); j++ {
			p := hull.GetIntAt(j, i)
			hullPoints = append(hullPoints, contour.At(int(p)))
		}
	}
	return hullPoints
}

func main() {
	// device ID
	webcam, err := gocv.OpenVideoCapture(0)
	handle(err)
	defer webcam.Close()

	window := gocv.NewWindow("Capture Window")
	defer window.Close()

	img := gocv.NewMat()
	blurred := gocv.NewMat()
	hsv := gocv.NewMat()
	mask := gocv.NewMat()
	hull := gocv.NewMat()
	defects := gocv.NewMat()
	defer img.Close()
	defer blurred.Close()
	defer hsv.Close()
	defer mask.Close()
	defer hull.Close()
	defer defects.Close()

	lower := gocv.NewScalar(0, 0, 100, 0)
	upper := gocv.NewScalar(30, 255, 255, 0)

	for {
		if ok := webcam.Read(&img); !ok {
			return
		}
		if img.Empty() {
			continue
		}

		// Mask
		gocv.GaussianBlur(img, &blurred, image.Pt(25, 25), 0, 0, gocv.BorderDefault)
		gocv.CvtColor(blurred, &hsv, gocv.ColorBGRToHSV)
		gocv.InRangeWithScalar(hsv, lower, upper, &mask)

		// Contours
		contours := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		ind := largestContour(contours)
		cnt := contours.At(ind)
		gocv.ConvexHull(cnt, &hull, false, false)
		gocv.ConvexityDefects(cnt, hull, &defects)

		// Drawing
		gocv.DrawContours(&img, contours, ind, color.RGBA{R: 0, G: 255, B: 255, A: 255}, 10)

		// Conversion inefficient
		hl := imPoint2Pts(hull, cnt)
		gocv.DrawContours(&img, gocv.NewPointsVectorFromPoints([][]image.Point{hl}), -1, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 10)

		// Defect processing
		var angle float64
		defectCount := 0
		for i := 0; i < defects.Rows(); i++ {
			start := cnt.At(int(defects.GetIntAt(i, 0)))
			end := cnt.At(int(defects.GetIntAt(i, 1)))
			far := cnt.At(int(defects.GetIntAt(i, 2)))

			a := math.Sqrt(math.Pow(float64(end.X-start.X), 2) + math.Pow(float64(end.Y-start.Y), 2))
			b := math.Sqrt(math.Pow(float64(far.X-start.X), 2) + math.Pow(float64(far.Y-start.Y), 2))
			c := math.Sqrt(math.Pow(float64(end.X-far.X), 2) + math.Pow(float64(end.Y-far.Y), 2))

			// apply cosine rule here
			angle = math.Acos((math.Pow(b, 2)+math.Pow(c, 2)-math.Pow(a, 2))/(2*b*c)) * 57

			// ignore angles > 90 and highlight rest with dots
			if angle <= 90 {
				defectCount++
				gocv.Circle(&img, far, 1, color.RGBA{R: 255, G: 0, B: 255, A: 255}, 2)
			}
		}
		gocv.PutText(&img, fmt.Sprintf("Fingers Up: %d", defectCount+1), image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2)

		window.IMShow(img)
		if window.WaitKey(1) == 27 {
			break
		}

		contours.Close()
	}
}
