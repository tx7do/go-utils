package math

import (
	"math"
)

// prop
//mean: the mean (μ) of the distribution
//variance: the variance (σ^2) of the distribution
//standardDeviation: the standard deviation (σ) of the distribution

// combination

type Gaussian struct {
	mean              float64
	variance          float64
	standardDeviation float64
}

func NewGaussian(mean, variance float64) *Gaussian {
	if variance <= 0.0 {
		panic("error")
	}

	return &Gaussian{
		mean:              mean,
		variance:          variance,
		standardDeviation: math.Sqrt(float64(variance)),
	}
}

// Erfc Complementary error function
// From Numerical Recipes in C 2e p221
func Erfc(x float64) float64 {
	z := math.Abs(x)
	t := 1 / (1 + z/2)
	r := t * math.Exp(-z*z-1.26551223+t*(1.00002368+
		t*(0.37409196+t*(0.09678418+t*(-0.18628806+
			t*(0.27886807+t*(-1.13520398+t*(1.48851587+
				t*(-0.82215223+t*0.17087277)))))))))
	if x >= 0 {
		return r
	} else {
		return 2 - r
	}
}

// Ierfc Inverse complementary error function
// From Numerical Recipes 3e p265
func Ierfc(x float64) float64 {
	if x >= 2 {
		return -100
	}
	if x <= 0 {
		return 100
	}
	var xx float64
	if x < 1 {
		xx = x
	} else {
		xx = 2 - x
	}
	t := math.Sqrt(-2 * math.Log(xx/2))
	r := -0.70711 * ((2.30753+t*0.27061)/
		(1+t*(0.99229+t*0.04481)) - t)

	for j := 0; j < 2; j++ {
		e := Erfc(r) - xx
		r += e / (1.12837916709551257*math.Exp(-(r*r)) - r*e)
	}

	if x < 1 {
		return r
	} else {
		return -r
	}

}

// fromPrecisionMean Construct a new distribution from the precision and precisionmean
func fromPrecisionMean(precision, precisionmean float64) *Gaussian {
	return NewGaussian(precisionmean/precision, 1/precision)
}

/// PROB

// Pdf pdf(x): the probability density function, which describes the probability
// of a random variable taking on the value x
func (g *Gaussian) Pdf(x float64) float64 {
	m := g.standardDeviation * math.Sqrt(2*math.Pi)
	e := math.Exp(-math.Pow(x-g.mean, 2) / (2 * g.variance))
	return e / m
}

// Cdf cdf(x): the cumulative distribution function,
// which describes the probability of a random
// variable falling in the interval (−∞, x]
func (g *Gaussian) Cdf(x float64) float64 {
	return 0.5 * Erfc(-(x-g.mean)/(g.standardDeviation*math.Sqrt(2)))
}

// Ppf ppf(x): the percent point function, the inverse of cdf
func (g *Gaussian) Ppf(x float64) float64 {
	return g.mean - g.standardDeviation*math.Sqrt(2)*Ierfc(2*x)
}

// Add add(d): returns the result of adding this and the given distribution
func (g *Gaussian) Add(d *Gaussian) *Gaussian {
	return NewGaussian(g.mean+d.mean, g.variance+d.variance)
}

// Sub sub(d): returns the result of subtracting this and the given distribution
func (g *Gaussian) Sub(d *Gaussian) *Gaussian {
	return NewGaussian(g.mean-d.mean, g.variance+d.variance)
}

// Scale scale(c): returns the result of scaling this distribution by the given constant
func (g *Gaussian) Scale(c float64) *Gaussian {
	return NewGaussian(g.mean*c, g.variance*c*c)
}

// Mul mul(d): returns the product distribution of this and the given distribution. If a constant is passed in the distribution is scaled.
func (g *Gaussian) Mul(d *Gaussian) *Gaussian {
	precision := 1 / g.variance
	dprecision := 1 / d.variance
	return fromPrecisionMean(precision+dprecision, precision*g.mean+dprecision*d.mean)
}

// Div div(d): returns the quotient distribution of this and the given distribution. If a constant is passed in the distribution is scaled by 1/d.
func (g *Gaussian) Div(d *Gaussian) *Gaussian {
	precision := 1 / g.variance
	dprecision := 1 / d.variance
	return fromPrecisionMean(precision-dprecision, precision*g.mean-dprecision*d.mean)
}
