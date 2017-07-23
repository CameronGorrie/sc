package sc

import (
	"math"
)

// C wraps a float32 and implements the Input interface.
type C float32

// Abs computes the absolute value of a signal.
func (c C) Abs() Input {
	return C(float32(math.Abs(float64(c))))
}

// Add adds another input to the constant.
func (c C) Add(val Input) Input {
	if v, ok := val.(C); ok {
		return C(float32(v) + float32(c))
	}
	return val.Add(c)
}

// AmpDb converts linear amplitude to decibels.
func (c C) AmpDb() Input {
	return C(float32(20 * math.Log10(float64(c))))
}

// Ceil computes the ceiling (next highest integer) of a signal.
func (c C) Ceil() Input {
	return C(float32(math.Ceil(float64(c))))
}

// Cpsmidi converts frequency in Hz to midi note values.
func (c C) Cpsmidi() Input {
	return C(Cpsmidi(float32(c)))
}

// Cpsoct converts cycles per second to decimal octaves.
func (c C) Cpsoct() Input {
	return C(Cpsoct(float32(c)))
}

// Cubed computes the cube of a signal.
func (c C) Cubed() Input {
	return C(float32(c * c * c))
}

// DbAmp converts decibels to linear amplitude.
func (c C) DbAmp() Input {
	return C(float32(math.Pow(10, float64(c)/20)))
}

// Exp computes the exponential of a signal.
func (c C) Exp() Input {
	return C(float32(math.Exp(float64(c))))
}

// Floor computes the floor of a constant.
func (c C) Floor() Input {
	return C(float32(math.Floor(float64(c))))
}

// Frac returns the fractional part of a constant.
func (c C) Frac() Input {
	return C(float32(float64(c) - math.Trunc(float64(c))))
}

// Max returns the maximum of one input and another.
func (c C) Max(other Input) Input {
	if v, ok := other.(C); ok {
		return C(maxFloat32(float32(c), float32(v)))
	}
	return other.Max(c)
}

// Midicps converts MIDI note number to cycles per second.
func (c C) Midicps() Input {
	return C(Midicps(float32(c)))
}

// Midiratio converts an interval in MIDI notes into a frequency ratio.
func (c C) Midiratio() Input {
	return C(float32(math.Pow(2, float64(c)/12)))
}

// Mul multiplies the constant by another input.
func (c C) Mul(val Input) Input {
	if v, ok := val.(C); ok {
		return C(float32(v) * float32(c))
	}
	return val.Mul(c)
}

// MulAdd multiplies and adds at the same time.
func (c C) MulAdd(mul, add Input) Input {
	if m, mok := mul.(C); mok {
		if a, aok := add.(C); aok {
			return C((float32(m) * float32(c)) + float32(a))
		}
		return add.MulAdd(c, mul)
	}
	return mul.MulAdd(c, add)
}

// Neg is a convenience operator that multiplies a signal by -1.
func (c C) Neg() Input {
	return C(float32(c) * -1)
}

// Octcps converts decimal octaves to cycles per second.
func (c C) Octcps() Input {
	return C(Octcps(float32(c)))
}

// Ratiomidi converts a frequency ratio to an interval in MIDI notes.
func (c C) Ratiomidi() Input {
	return C(float32(12 * math.Log2(float64(c))))
}

// Reciprocal computes the reciprocal of a signal.
func (c C) Reciprocal() Input {
	return C(1 / float32(c))
}

// Sign computes the sign of the constant.
func (c C) Sign() Input {
	if c > 0 {
		return C(1)
	} else if c < 0 {
		return C(-1)
	}
	return C(0)
}

// SoftClip clips the constant to the range [-0.5, 0.5]
func (c C) SoftClip() Input {
	if float32(c) < -0.5 {
		return C(-0.5)
	} else if float32(c) > 0.5 {
		return C(0.5)
	}
	return c
}

// Sqrt computes the square root of a constant.
func (c C) Sqrt() Input {
	return C(math.Sqrt(float64(c)))
}

// Squared computes the square of a signal.
func (c C) Squared() Input {
	return C(float32(c * c))
}

func maxFloat32(f1, f2 float32) float32 {
	if f1 > f2 {
		return f1
	}
	return f2
}
