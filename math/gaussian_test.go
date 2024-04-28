package math

import (
	"fmt"
	"testing"
)

func TestGaussian(t *testing.T) {
	g := NewGaussian(3.0, 1)

	fmt.Printf("g: %#v\n", g)
	fmt.Printf("pdf: %f\n", g.Pdf(5))
	fmt.Printf("cdf: %f\n", g.Cdf(2))
	fmt.Printf("ppf: %f\n", g.Ppf(5))

	d := NewGaussian(0, 1)
	fmt.Printf("ppf: %f, %f\n", d.Pdf(-2), 0.053991)
	fmt.Printf("ppf: %f, %f\n", d.Pdf(-1), 0.241971)
	fmt.Printf("ppf: %f, %f\n", d.Pdf(0), 0.398942)
	fmt.Printf("ppf: %f, %f\n", d.Pdf(1), 0.241971)
	fmt.Printf("ppf: %f, %f\n", d.Pdf(2), 0.053991)

	fmt.Printf("cdf: %f, %f\n", d.Cdf(-1.28155), 0.1)
	fmt.Printf("cdf: %f, %f\n", d.Cdf(-0.67499), 0.25)
	fmt.Printf("cdf: %f, %f\n", d.Cdf(0), 0.5)
	fmt.Printf("cdf: %f, %f\n", d.Cdf(0.67499), 0.75)
	fmt.Printf("cdf: %f, %f\n", d.Cdf(1.28155), 0.9)

	fmt.Printf("ppf: %f, %f\n", d.Ppf(0.1), -1.28155)
	fmt.Printf("ppf: %f, %f\n", d.Ppf(0.25), -0.67499)
	fmt.Printf("ppf: %f, %f\n", d.Ppf(0.5), 0.0)
	fmt.Printf("ppf: %f, %f\n", d.Ppf(0.75), 0.67449)
	fmt.Printf("ppf: %f, %f\n", d.Ppf(0.9), 1.28155)

	d = d.Mul(NewGaussian(0, 1))
	fmt.Printf("Mul: %#v\n", d)
	fmt.Printf("%#v\n%#v", NewGaussian(1, 1).Scale(2), NewGaussian(2, 4))

	d = NewGaussian(1, 1).Div(NewGaussian(1, 2))
	fmt.Printf("div\n")
	fmt.Printf("%#v\n%#v\n", d, NewGaussian(1, 2))
	fmt.Printf("%#v\n%#v\n", NewGaussian(1, 1).Scale(1/(1.0/2.0)), NewGaussian(2, 4))

	fmt.Printf("ADD:\n%#v\n%#v\n", NewGaussian(1, 1).Add(NewGaussian(1, 2)), NewGaussian(2, 3))
	fmt.Printf("SUB:\n%#v\n%#v\n", NewGaussian(1, 1).Sub(NewGaussian(1, 2)), NewGaussian(0, 3))
	fmt.Printf("SCALE:\n%#v\n%#v\n", NewGaussian(1, 1).Scale(2), NewGaussian(2, 4))
}
