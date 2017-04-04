package core

import (
    "log"
    "math"
)

type Vector struct {
    Length int
    Values []float64
}

func NewVector(length int) *Vector {
    vector := new(Vector)
    vector.Length = length
    vector.Values = make([]float64, length)

    return vector
}

func DeepCopy(vector *Vector) *Vector {
    v := NewVector(vector.Length)
    for i := 0; i < vector.Length; i++ {
        v.Values[i] = vector.Values[i]
    }
    return v
}

func (v *Vector) Sum(vector1 *Vector) {
    if v.Length != vector1.Length{
        log.Fatal("v, vector1")
    }

    for i := 0; i < vector1.Length; i++ {
        v.Values[i] = v.Values[i] + vector1.Values[i]
    }
}


func (v *Vector) Multiply(a float64) {
    for i := 0; i < v.Length; i++ {
        v.Values[i] = v.Values[i] * a
    }
}

func (v *Vector) Norm() {
    result := 0.0
    for i := 0; i < v.Length; i++ {
        result += v.Values[i] * v.Values[i]
    }
    result = math.Sqrt(result)
    for i := 0; i < v.Length; i++ {
        v.Values[i] /= result
    }
}


// 计算点乘积 vector1^T * vector2
func VecDotProduct(vector1, vector2 *Vector) float64 {
    if vector1.Length != vector2.Length {
        log.Fatal("vector1和vector2的长度不一致")
    }

    var result float64
    result = 0
    for iterVec1 := 0; iterVec1 < vector1.Length; iterVec1++ {
        result += vector1.Values[iterVec1] * vector2.Values[iterVec1]
    }
    return result
}


func (v *Vector) Increment(vector1 *Vector, g float64) {
    if v.Length != vector1.Length {
        log.Fatal("vector1和v的长度不一致")
    }

    for i := 0; i < v.Length; i++ {
        v.Values[i] += g*vector1.Values[i]
    }
}
