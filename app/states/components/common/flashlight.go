package common

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const DefaultFlashlightSize = 168.0
const DefaultFlashlightDuration = 800.0

type Flashlight struct {
	flShader   *shader.RShader
	vao        *buffer.VertexArrayObject
	time       float64
	delta      float64
	position   vector.Vector2f
	size       *animation.Glider
	dim        *animation.Glider
	beatMap    *beatmap.BeatMap
	breakIndex int
	target     float64
	sliding    bool
}

func NewFlashlight(beatMap *beatmap.BeatMap) *Flashlight {
	vert, err := assets.GetString("assets/shaders/flashlight.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/flashlight.fsh")
	if err != nil {
		panic(err)
	}

	flShader := shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	vao := buffer.NewVertexArrayObject()
	vao.AddVBO("default", 6, 0, attribute.Format{
		attribute.VertexAttribute{Name: "in_position", Type: attribute.Vec2},
	})

	vao.SetData("default", 0, []float32{
		-1, -1,
		1, -1,
		1, 1,
		1, 1,
		-1, 1,
		-1, -1,
	})

	vao.Attach(flShader)

	size := animation.NewGlider(DefaultFlashlightSize * 2.5)

	startTime := float64(beatMap.HitObjects[0].GetStartTime())
	endTime := float64(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime()) + float64(beatMap.Diff.Hit50+5)

	size.AddEvent(startTime-DefaultFlashlightDuration, startTime, DefaultFlashlightSize)
	size.AddEvent(endTime, endTime+DefaultFlashlightDuration, DefaultFlashlightSize*2.5)

	return &Flashlight{
		flShader:   flShader,
		vao:        vao,
		size:       size,
		beatMap:    beatMap,
		breakIndex: -1,
		dim:        animation.NewGlider(0.0),
	}
}

func (fl *Flashlight) UpdatePosition(cursorPosition vector.Vector2f) {
	oldPosition := fl.position
	fl.position = cursorPosition.Sub(oldPosition).Scl(float32(easing.OutQuad(math.Min(fl.delta, 120) / 120))).Add(oldPosition)
}

func (fl *Flashlight) UpdateCombo(combo int64) {
	target := DefaultFlashlightSize

	switch {
	case combo > 200:
		target *= 0.625
	case combo > 100:
		target *= 0.8125
	}

	fl.target = target

	fl.size.AddEvent(fl.time, fl.time+DefaultFlashlightDuration, target)
}

func (fl *Flashlight) SetSliding(value bool) {
	if fl.sliding != value {
		dim := 0.0
		if value {
			dim = 0.8
		}

		fl.dim.AddEvent(fl.time, fl.time+50, dim)

		fl.sliding = value
	}
}

func (fl *Flashlight) Update(time float64) {
	fl.delta = time - fl.time

	fl.time = time

	for i := fl.breakIndex + 1; i < len(fl.beatMap.Pauses); i++ {
		pause := fl.beatMap.Pauses[i]

		if time < float64(pause.GetStartTime()) {
			break
		}

		fl.breakIndex = i

		if float64(pause.EndTime-pause.StartTime) > DefaultFlashlightDuration*2 {
			fl.size.AddEvent(float64(pause.StartTime), float64(pause.StartTime)+DefaultFlashlightDuration, DefaultFlashlightSize*2.5)
			fl.size.AddEvent(float64(pause.EndTime)-DefaultFlashlightDuration, float64(pause.EndTime), fl.target)
		}
	}

	fl.size.Update(time)
	fl.dim.Update(time)
}

func (fl *Flashlight) Draw(matrix mgl32.Mat4) {
	blend.Push()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.One)

	fl.flShader.Bind()
	fl.flShader.SetUniform("cursorPosition", mgl32.Vec2{fl.position.X, fl.position.Y})
	fl.flShader.SetUniform("radius", float32(fl.size.GetValue()))
	fl.flShader.SetUniform("dim", float32(fl.dim.GetValue()))
	fl.flShader.SetUniform("invMatrix", matrix.Inv())

	fl.vao.Bind()
	fl.vao.Draw()
	fl.vao.Unbind()

	fl.flShader.Unbind()

	blend.Pop()
}