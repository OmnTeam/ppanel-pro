package captcha

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	sliderBgWidth     = 560
	sliderBgHeight    = 280
	sliderBlockSize   = 100
	sliderMinX        = 140
	sliderMaxX        = 420
	sliderTolerance   = 6
	sliderExpiry      = 5 * time.Minute
	sliderTokenExpiry = 30 * time.Second
)

type sliderShape int

const (
	shapeSquare sliderShape = iota
	shapeCircle
	shapeDiamond
	shapeStar
	shapeTriangle
	shapeTrapezoid
)

type sliderService struct {
	redis *redis.Client
}

type sliderData struct {
	X     int         `json:"x"`
	Y     int         `json:"y"`
	Shape sliderShape `json:"shape"`
}

type TrailPoint struct {
	X int   `json:"x"`
	Y int   `json:"y"`
	T int64 `json:"t"`
}

func newSliderService(redisClient *redis.Client) *sliderService {
	return &sliderService{redis: redisClient}
}

func (s *sliderService) GenerateSlider(ctx context.Context) (id string, bgImage string, blockImage string, err error) {
	if s.redis == nil {
		return "", "", "", fmt.Errorf("redis client is nil")
	}

	bg := generateBackground()
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	x := sliderMinX + randomizer.Intn(sliderMaxX-sliderMinX)
	y := randomizer.Intn(sliderBgHeight - sliderBlockSize)
	shape := sliderShape(randomizer.Intn(6))

	block := cropBlockShaped(bg, x, y, shape)
	cutBackgroundShaped(bg, x, y, shape)

	bgB64, err := imageToPNGBase64(bg)
	if err != nil {
		return "", "", "", err
	}
	blockB64, err := imageToPNGBase64(block)
	if err != nil {
		return "", "", "", err
	}

	id = uuid.NewString()
	payload, _ := json.Marshal(sliderData{X: x, Y: y, Shape: shape})
	if err := s.redis.Set(ctx, sliderCacheKey(id), payload, sliderExpiry).Err(); err != nil {
		return "", "", "", err
	}

	return id, bgB64, blockB64, nil
}

func (s *sliderService) Generate(ctx context.Context) (id string, image string, err error) {
	id, _, _, err = s.GenerateSlider(ctx)
	return id, "", err
}

func (s *sliderService) VerifySlider(ctx context.Context, id string, x, y int, trail string) (string, error) {
	if trail == "" {
		return "", fmt.Errorf("trail required")
	}

	var points []TrailPoint
	if err := json.Unmarshal([]byte(trail), &points); err != nil {
		return "", fmt.Errorf("invalid trail")
	}
	if !validateTrail(points, x) {
		return "", fmt.Errorf("trail validation failed")
	}

	value, err := s.redis.Get(ctx, sliderCacheKey(id)).Result()
	if err != nil {
		return "", fmt.Errorf("captcha not found or expired")
	}

	var data sliderData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return "", fmt.Errorf("invalid captcha data")
	}

	if abs(x-data.X) > sliderTolerance || abs(y-data.Y) > sliderTolerance {
		_ = s.redis.Del(ctx, sliderCacheKey(id)).Err()
		return "", fmt.Errorf("position mismatch")
	}

	_ = s.redis.Del(ctx, sliderCacheKey(id)).Err()

	token := uuid.NewString()
	if err := s.redis.Set(ctx, sliderTokenCacheKey(token), "1", sliderTokenExpiry).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *sliderService) VerifySliderToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	value, err := s.redis.Get(ctx, sliderTokenCacheKey(token)).Result()
	if err != nil || value != "1" {
		return false, nil
	}
	_ = s.redis.Del(ctx, sliderTokenCacheKey(token)).Err()
	return true, nil
}

func (s *sliderService) Verify(ctx context.Context, token string, code string, ip string) (bool, error) {
	return s.VerifySliderToken(ctx, token)
}

func (s *sliderService) GetType() CaptchaType {
	return CaptchaTypeSlider
}

func sliderCacheKey(id string) string {
	return fmt.Sprintf("captcha:slider:%s", id)
}

func sliderTokenCacheKey(token string) string {
	return fmt.Sprintf("captcha:slider:token:%s", token)
}

func validateTrail(trail []TrailPoint, declaredX int) bool {
	if len(trail) < 8 {
		return false
	}

	duration := trail[len(trail)-1].T - trail[0].T
	if duration < 300 || duration > 15000 {
		return false
	}
	if trail[0].X > 10 {
		return false
	}

	var speeds []float64
	for i := 1; i < len(trail); i++ {
		dt := float64(trail[i].T - trail[i-1].T)
		dx := float64(trail[i].X - trail[i-1].X)
		dy := float64(trail[i].Y - trail[i-1].Y)
		if abs(int(dx)) > 80 {
			return false
		}
		if dt > 0 {
			dist := math.Sqrt(dx*dx + dy*dy)
			speeds = append(speeds, dist/dt)
		}
	}

	if len(speeds) >= 3 {
		mean := 0.0
		for _, v := range speeds {
			mean += v
		}
		mean /= float64(len(speeds))

		variance := 0.0
		for _, v := range speeds {
			d := v - mean
			variance += d * d
		}
		variance /= float64(len(speeds))
		if variance < 1e-6 {
			return false
		}
	}

	minY := trail[0].Y
	maxY := trail[0].Y
	for _, point := range trail {
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	if maxY-minY < 2 {
		return false
	}

	if abs(trail[len(trail)-1].X-declaredX) > sliderTolerance*2 {
		return false
	}

	return true
}

func inMask(dx, dy int, shape sliderShape) bool {
	half := sliderBlockSize / 2
	switch shape {
	case shapeCircle:
		ex := dx - half
		ey := dy - half
		return ex*ex+ey*ey <= half*half
	case shapeDiamond:
		return abs(dx-half)+abs(dy-half) <= half
	case shapeStar:
		return inStar(dx, dy, half)
	case shapeTriangle:
		return inTriangle(dx, dy)
	case shapeTrapezoid:
		return inTrapezoid(dx, dy)
	default:
		margin := 8
		return dx >= margin && dx < sliderBlockSize-margin && dy >= margin && dy < sliderBlockSize-margin
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func pointInPolygon(x, y float64, points [][2]float64) bool {
	inside := false
	j := len(points) - 1
	for i := 0; i < len(points); i++ {
		xi, yi := points[i][0], points[i][1]
		xj, yj := points[j][0], points[j][1]
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func inStar(dx, dy, half int) bool {
	cx, cy := float64(half), float64(half)
	outer := float64(half) * 0.92
	inner := float64(half) * 0.40
	x := float64(dx) - cx
	y := float64(dy) - cy

	points := make([][2]float64, 10)
	for i := 0; i < 10; i++ {
		angle := float64(i)*math.Pi/5 - math.Pi/2
		radius := outer
		if i%2 == 1 {
			radius = inner
		}
		points[i] = [2]float64{radius * math.Cos(angle), radius * math.Sin(angle)}
	}

	return pointInPolygon(x, y, points)
}

func inTriangle(dx, dy int) bool {
	margin := 5
	size := sliderBlockSize - 2*margin
	half := float64(sliderBlockSize) / 2

	ax, ay := half, float64(margin)
	bx, by := float64(margin), float64(margin+size)
	cx, cy := float64(margin+size), float64(margin+size)
	px, py := float64(dx), float64(dy)

	d1 := (px-bx)*(ay-by) - (ax-bx)*(py-by)
	d2 := (px-cx)*(by-cy) - (bx-cx)*(py-cy)
	d3 := (px-ax)*(cy-ay) - (cx-ax)*(py-ay)
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

func inTrapezoid(dx, dy int) bool {
	margin := 5
	topY := float64(margin)
	bottomY := float64(sliderBlockSize - margin)
	totalHeight := bottomY - topY
	half := float64(sliderBlockSize) / 2
	topHalfWidth := float64(sliderBlockSize) * 0.25
	bottomHalfWidth := float64(sliderBlockSize) * 0.45
	x, y := float64(dx), float64(dy)

	if y < topY || y > bottomY {
		return false
	}

	ratio := (y - topY) / totalHeight
	halfWidth := topHalfWidth + ratio*(bottomHalfWidth-topHalfWidth)
	return x >= half-halfWidth && x <= half+halfWidth
}

func cropBlockShaped(bg *image.NRGBA, x, y int, shape sliderShape) *image.NRGBA {
	block := image.NewNRGBA(image.Rect(0, 0, sliderBlockSize, sliderBlockSize))
	for dy := 0; dy < sliderBlockSize; dy++ {
		for dx := 0; dx < sliderBlockSize; dx++ {
			if inMask(dx, dy, shape) {
				block.SetNRGBA(dx, dy, bg.NRGBAAt(x+dx, y+dy))
			}
		}
	}

	borderColor := color.NRGBA{R: 255, G: 255, B: 255, A: 230}
	for dy := 0; dy < sliderBlockSize; dy++ {
		for dx := 0; dx < sliderBlockSize; dx++ {
			if !inMask(dx, dy, shape) {
				continue
			}

			nearEdge := false
		check:
			for ddy := -2; ddy <= 2; ddy++ {
				for ddx := -2; ddx <= 2; ddx++ {
					if abs(ddx)+abs(ddy) > 2 {
						continue
					}
					nx, ny := dx+ddx, dy+ddy
					if nx < 0 || nx >= sliderBlockSize || ny < 0 || ny >= sliderBlockSize || !inMask(nx, ny, shape) {
						nearEdge = true
						break check
					}
				}
			}
			if nearEdge {
				block.SetNRGBA(dx, dy, borderColor)
			}
		}
	}

	return block
}

func cutBackgroundShaped(bg *image.NRGBA, x, y int, shape sliderShape) {
	holeColor := color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	borderColor := color.NRGBA{R: 255, G: 255, B: 255, A: 220}

	for dy := 0; dy < sliderBlockSize; dy++ {
		for dx := 0; dx < sliderBlockSize; dx++ {
			if inMask(dx, dy, shape) {
				bg.SetNRGBA(x+dx, y+dy, holeColor)
			}
		}
	}

	for dy := 0; dy < sliderBlockSize; dy++ {
		for dx := 0; dx < sliderBlockSize; dx++ {
			if !inMask(dx, dy, shape) {
				continue
			}

			nearEdge := false
		check:
			for ddy := -2; ddy <= 2; ddy++ {
				for ddx := -2; ddx <= 2; ddx++ {
					if abs(ddx)+abs(ddy) > 2 {
						continue
					}
					nx, ny := dx+ddx, dy+ddy
					if nx < 0 || nx >= sliderBlockSize || ny < 0 || ny >= sliderBlockSize || !inMask(nx, ny, shape) {
						nearEdge = true
						break check
					}
				}
			}
			if nearEdge {
				bg.SetNRGBA(x+dx, y+dy, borderColor)
			}
		}
	}
}

func generateBackground() *image.NRGBA {
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	img := image.NewNRGBA(image.Rect(0, 0, sliderBgWidth, sliderBgHeight))

	blockWidth := 60
	blockHeight := 60
	palette := []color.NRGBA{
		{R: 70, G: 130, B: 180, A: 255},
		{R: 60, G: 179, B: 113, A: 255},
		{R: 205, G: 92, B: 92, A: 255},
		{R: 255, G: 165, B: 0, A: 255},
		{R: 147, G: 112, B: 219, A: 255},
		{R: 64, G: 224, B: 208, A: 255},
		{R: 220, G: 120, B: 60, A: 255},
		{R: 100, G: 149, B: 237, A: 255},
	}

	for by := 0; by*blockHeight < sliderBgHeight; by++ {
		for bx := 0; bx*blockWidth < sliderBgWidth; bx++ {
			base := palette[randomizer.Intn(len(palette))]
			x0 := bx * blockWidth
			y0 := by * blockHeight
			x1 := x0 + blockWidth
			y1 := y0 + blockHeight

			for py := y0; py < y1 && py < sliderBgHeight; py++ {
				for px := x0; px < x1 && px < sliderBgWidth; px++ {
					variation := int8(randomizer.Intn(41) - 20)
					img.SetNRGBA(px, py, color.NRGBA{
						R: addVariation(base.R, variation),
						G: addVariation(base.G, variation),
						B: addVariation(base.B, variation),
						A: 255,
					})
				}
			}
		}
	}

	numCircles := 6 + randomizer.Intn(6)
	for i := 0; i < numCircles; i++ {
		cx := randomizer.Intn(sliderBgWidth)
		cy := randomizer.Intn(sliderBgHeight)
		radius := 18 + randomizer.Intn(30)
		drawCircle(img, cx, cy, radius, color.NRGBA{
			R: uint8(randomizer.Intn(256)),
			G: uint8(randomizer.Intn(256)),
			B: uint8(randomizer.Intn(256)),
			A: 180,
		})
	}

	return img
}

func addVariation(base uint8, variation int8) uint8 {
	value := int(base) + int(variation)
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

func drawCircle(img *image.NRGBA, cx, cy, radius int, fill color.NRGBA) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= radius*radius &&
				x >= bounds.Min.X && x < bounds.Max.X &&
				y >= bounds.Min.Y && y < bounds.Max.Y {
				img.SetNRGBA(x, y, fill)
			}
		}
	}
}

func imageToPNGBase64(img image.Image) (string, error) {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}
