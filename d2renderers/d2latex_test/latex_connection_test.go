package d2latex_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2exporter"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2renderers/d2fonts"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

func TestRenderLatexConnectionLabel(t *testing.T) {
	rootObject := &d2graph.Object{
		ID: "root",
		Attributes: d2graph.Attributes{
			Label: d2graph.Scalar{
				Value: "Root",
			},
		},
	}

	srcObject := &d2graph.Object{
		ID: "source",
		Attributes: d2graph.Attributes{
			Label: d2graph.Scalar{
				Value: "Source",
			},
		},
		Parent: rootObject,
	}

	dstObject := &d2graph.Object{
		ID: "destination",
		Attributes: d2graph.Attributes{
			Label: d2graph.Scalar{
				Value: "Destination",
			},
		},
		Parent: rootObject,
	}

	label := `e^{i\pi}+1=0`
	graph := &d2graph.Graph{
		Root: rootObject,
		Edges: []*d2graph.Edge{
			{
				Attributes: d2graph.Attributes{
					Label: d2graph.Scalar{
						Value: label,
					},
				},
				IsLatex: true,
				Route: []*geo.Point{
					{X: 20, Y: -20},
					{X: 20, Y: 80},
				},
				Src:             srcObject,
				Dst:             dstObject,
				LabelPosition:   go2.Pointer("INSIDE_MIDDLE_CENTER"),
				LabelPercentage: go2.Pointer(0.5),
			},
		},
	}

	fontFamily := d2fonts.SourceSansPro
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatalf("failed to create ruler: %v", err)
	}

	err = graph.SetDimensions(nil, ruler, &fontFamily)
	if err != nil {
		t.Fatalf("failed to set dimensions: %v", err)
	}

	diagram, err := d2exporter.Export(context.Background(), graph, &fontFamily)
	if err != nil {
		t.Fatalf("failed to export graph: %v", err)
	}

	opts := &d2svg.RenderOpts{
		ThemeID: go2.Pointer(d2themescatalog.NeutralDefault.ID),
	}

	svgBytes, err := d2svg.Render(diagram, opts)
	if err != nil {
		t.Fatalf("failed to render connection: %v", err)
	}

	filePath := "test_connection_label.svg"
	err = os.WriteFile(filePath, svgBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write SVG to file: %v", err)
	}
	if !strings.Contains(string(svgBytes), `data-mml-node="math"`) {
		t.Errorf("expected LaTeX-rendered SVG in connection label, got: %s", string(svgBytes))
	}
}
