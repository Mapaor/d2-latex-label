package d2latex

import (
	_ "embed"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"oss.terrastruct.com/d2/lib/jsrunner"
	"oss.terrastruct.com/util-go/xdefer"
)

var pxPerEx = 8

//go:embed polyfills.js
var polyfillsJS string

//go:embed setup.js
var setupJS string

//go:embed mathjax.js
var mathjaxJS string

// Matches this
// <svg style="background: white;" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="563" height="326" viewBox="-100 -100 563 326"><style type="text/css">
var svgRe = regexp.MustCompile(`<svg[^>]+width="([0-9\.]+)ex" height="([0-9\.]+)ex"[^>]+>`)

func Render(s string) (_ string, err error) {
	defer xdefer.Errorf(&err, "latex failed to parse")
	fmt.Printf("Rendering LaTeX string: %s\n", s)

	// Escape backslashes in the LaTeX string
	s = doubleBackslashes(s)
	fmt.Printf("Escaped LaTeX string: %s\n", s)

	// Initialize the JavaScript runner
	runner := jsrunner.NewJSRunner()

	// Run polyfills.js
	if _, err := runner.RunString(polyfillsJS); err != nil {
		fmt.Printf("Error running polyfillsJS: %v\n", err)
		return "", err
	}

	// Run mathjax.js
	if _, err := runner.RunString(mathjaxJS); err != nil {
		fmt.Printf("Error running mathjaxJS: %v\n", err)
		// Known issue that a harmless error occurs in JS: https://github.com/mathjax/MathJax/issues/3289
		if runner.Engine() == jsrunner.Goja {
			return "", err
		}
	}

	// Run setup.js
	if _, err := runner.RunString(setupJS); err != nil {
		fmt.Printf("Error running setupJS: %v\n", err)
		return "", err
	}

	// Run the LaTeX conversion
	val, err := runner.RunString(fmt.Sprintf(`adaptor.innerHTML(html.convert(`+"`"+"%s`"+`, {
      em: %d,
      ex: %d,
    }))`, s, pxPerEx*2, pxPerEx))
	if err != nil {
		fmt.Printf("Error running LaTeX conversion: %v\n", err)
		return "", err
	}

	return val.String(), nil
}

func Measure(s string) (width, height int, err error) {
	defer xdefer.Errorf(&err, "latex failed to parse")
	fmt.Printf("Measuring LaTeX string: %s\n", s)

	svg, err := Render(s)
	if err != nil {
		fmt.Printf("Error rendering LaTeX string: %v\n", err)
		return 0, 0, err
	}

	fmt.Printf("Rendered SVG: %s\n", svg)

	dims := svgRe.FindAllStringSubmatch(svg, -1)
	if len(dims) != 1 || len(dims[0]) != 3 {
		fmt.Printf("SVG parsing failed for LaTeX: %v\n", svg)
		return 0, 0, fmt.Errorf("svg parsing failed for latex: %v", svg)
	}

	wEx := dims[0][1]
	hEx := dims[0][2]

	wf, err := strconv.ParseFloat(wEx, 64)
	if err != nil {
		fmt.Printf("Error parsing width from SVG: %v\n", err)
		return 0, 0, fmt.Errorf("svg parsing failed for latex: %v", svg)
	}
	hf, err := strconv.ParseFloat(hEx, 64)
	if err != nil {
		fmt.Printf("Error parsing height from SVG: %v\n", err)
		return 0, 0, fmt.Errorf("svg parsing failed for latex: %v", svg)
	}

	width = int(math.Ceil(wf * float64(pxPerEx)))
	height = int(math.Ceil(hf * float64(pxPerEx)))

	fmt.Printf("Measured dimensions: Width=%d, Height=%d\n", width, height)
	return width, height, nil
}

func doubleBackslashes(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			result.WriteString("\\\\")
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
