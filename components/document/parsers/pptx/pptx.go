package pptx

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image"
	"regexp"
	"strings"

	qxml "github.com/dgrr/quickxml"

	"github.com/bububa/atomic-agents/components/document/parsers"
)

// PPTx represents the XML file structure and settings for parsing a pptx file.
type PPTx struct {
	zipReadCloser *zip.ReadCloser
	slideFiles    map[int]*zip.File
	chartsFiles   map[string]*zip.File
	imagesFiles   map[string]*zip.File
	diagramsFiles map[string]*zip.File
	slideRelsMap  map[int]map[string]string

	parseCharts   bool
	parseImages   bool
	parseDiagrams bool
	drawingsNoFmt bool
	ocr           parsers.OCR

	slideSep     string
	paragraphSep string
	phraseSep    string
	tableRowSep  string
	tableColSep  string
}

func NewPPTx() *PPTx {
	return &PPTx{
		slideSep:     strings.Repeat("-", 100) + "\n",
		paragraphSep: "\n",
		phraseSep:    " ",
		tableRowSep:  "\n",
		tableColSep:  "\t",
	}
}

// Close closes the zipReader and OCR client.
// After extracting the text, please remember to call this method.
func (pp *PPTx) Close() (err error) {
	if pp.zipReadCloser != nil {
		err = pp.zipReadCloser.Close()
		if err != nil {
			return
		}
	}
	if pp.ocr != nil {
		err = pp.ocr.Close()
		if err != nil {
			return
		}
	}

	return nil
}

// SetSlideSep sets slide text separator. Default is "-"x100.
func (pp *PPTx) SetSlideSep(sep string) {
	pp.slideSep = sep
}

// SetParagraphSep sets phrase separator. Default is " ".
func (pp *PPTx) SetPhraseSep(sep string) {
	pp.phraseSep = sep
}

// SetTableRowSep sets table row separator. Default is "\n".
func (pp *PPTx) SetTableRowSep(sep string) {
	pp.tableRowSep = sep
}

// SetTableColSep sets table column separator. Default is "\t".
func (pp *PPTx) SetTableColSep(sep string) {
	pp.tableColSep = sep
}

// SetParseCharts parses charts or not. Default is false.
func (pp *PPTx) SetParseCharts(v bool) {
	pp.parseCharts = v
}

// SetParseDiagrams parses diagrams or not. Default is false.
func (pp *PPTx) SetParseDiagrams(v bool) {
	pp.parseDiagrams = v
}

// SetParseImages parses images or not. Default is false.
// When ocr interface is not set, default tesseract-ocr will be used.
func (pp *PPTx) SetParseImages(v bool) {
	pp.parseImages = v
}

// SetDrawingsNoFmt sets drawings text no outline format.
func (pp *PPTx) SetDrawingsNoFmt(v bool) {
	pp.drawingsNoFmt = v
}

// SetOcrInterface overrides default ocr interface.
func (pp *PPTx) SetOcrInterface(ocr parsers.OCR) {
	pp.ocr = ocr
}

// NumSlides returns the number of slides.
func (pp *PPTx) NumSlides() int {
	return len(pp.slideFiles)
}

// ExtractImages extracts images from the pptx file.
//
// Parameters:
//   - None
//
// Returns:
//   - []types.Image: a slice of images extracted from the pptx file.
//   - error: an error if any occurred during the extraction process.
func (pp *PPTx) ExtractImages() ([]Image, error) {
	images := make([]Image, 0, len(pp.imagesFiles))
	for name, f := range pp.imagesFiles {
		r, err := f.Open()
		if err != nil {
			return images, err
		}
		img, format, err := image.Decode(r)
		if err != nil {
			r.Close()
			return images, err
		}
		r.Close()

		images = append(images, Image{
			Raw:    img,
			Name:   name,
			Format: format,
		})
	}

	return images, nil
}

// ExtractSlideTexts extracts the texts from the specified pptx slides(start 1).
//
// It takes in one or more slide numbers as parameters and returns a string
// containing the extracted texts. The function also returns an error if there
// is any issue with parsing the slides.
//
// Parameters:
//   - slides: An integer slice containing the slide numbers to extract texts from.
//
// Returns:
//   - string: A string containing the extracted texts.
//   - error: An error object if there is any issue with parsing the slides.
func (pp *PPTx) ExtractSlideTexts(slides ...int) (string, error) {
	texts := new(strings.Builder)
	for _, slide := range slides {
		slide, err := pp.parseSlide(slide)
		if err != nil {
			return texts.String(), err
		}
		if slide.Len() > 0 {
			texts.WriteString(slide.String())
			texts.WriteString(pp.slideSep)
		}
	}

	return texts.String(), nil
}

// ExtractTexts extracts the texts from the pptx file.
//
// It iterates through each slide of the pptx file and appends the text content
// to a strings.Builder object. The extracted texts are then returned as a string.
// If there is an error encountered during the parsing of a slide, the function
// returns the extracted texts up to that point, along with the error.
//
// Returns:
//   - string: The extracted texts from the pptx file.
//   - error: An error, if any, encountered during the parsing of the slides.
func (pp *PPTx) ExtractTexts() (string, error) {
	texts := new(strings.Builder)

	for i := 1; i <= pp.NumSlides(); i++ {
		slide, err := pp.parseSlide(i)
		if err != nil {
			return texts.String(), err
		}
		if slide.Len() > 0 {
			texts.WriteString(slide.String())
			texts.WriteString(pp.slideSep)
		}
	}

	return texts.String(), nil
}

// parseSlide parses a slide at the given index and returns the extracted texts, tables, charts, diagrams, and images.
//
// Parameters:
//   - i: the index of the slide to parse.
//
// Returns:
//   - texts: a strings.Builder containing the extracted texts.
//   - error: an error if the slide does not exist or if there was an error opening the slide file.
func (pp *PPTx) parseSlide(i int) (*strings.Builder, error) {
	slideFile, ok := pp.slideFiles[i]
	if !ok {
		return nil, fmt.Errorf("no slide exists at index %d", i)
	}

	rc, err := slideFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var (
		texts  = new(strings.Builder)
		phrase = ""
	)
	r := qxml.NewReader(rc)

NEXT:
	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.EndElement:
			if e.Name() == "a:p" {
				texts.WriteString(pp.paragraphSep)
			}

		case *qxml.StartElement:
			switch e.Name() {
			case "a:t":
				r.AssignNext(&phrase)
				if !r.Next() {
					break NEXT
				}
				if len(phrase) > 0 {
					texts.WriteString(phrase)
					texts.WriteString(pp.phraseSep)
					phrase = ""
				}

			case "a:tbl":
				table := pp.extractTable(r)
				if table != nil {
					texts.WriteString(table.String())
				}

			case "c:chart":
				if !pp.parseCharts {
					continue
				}
				attrs := e.Attrs()
				if attrs.Len() > 0 {
					rIdKV := attrs.Get("r:id")
					chart, _ := pp.extractChart(i, rIdKV.Value())
					if chart != nil {
						texts.WriteString(chart.String())
					}
				}

			case "dgm:relIds":
				if !pp.parseDiagrams {
					continue
				}
				attrs := e.Attrs()
				if attrs.Len() > 0 {
					rIdKV := attrs.Get("r:dm")
					diagram, _ := pp.extractDiagram(i, rIdKV.Value())
					if diagram != nil {
						texts.WriteString(diagram.String())
					}
				}

			case "a:blip":
				if !pp.parseImages {
					continue
				}
				attrs := e.Attrs()
				if attrs.Len() > 0 {
					rIdKV := attrs.Get("r:embed")
					image, _ := pp.extractImage(i, rIdKV.Value())
					if image != nil {
						texts.WriteString(image.String())
					}
				}
			}
		}
	}
	return texts, nil
}

// extractTable extracts table data from a pptx file using a qxml.Reader.
//
// Parameters:
//   - r: a qxml.Reader object used to read the XML elements.
//
// Return type:
//   - *strings.Builder: a strings.Builder object containing the extracted table data.
func (pp *PPTx) extractTable(r *qxml.Reader) *strings.Builder {
	var (
		texts = new(strings.Builder)
		row   = new(strings.Builder)
		a_t   = ""
	)

NEXT:
	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.StartElement:
			if e.Name() == "a:t" {
				r.AssignNext(&a_t)
				if !r.Next() {
					break NEXT
				}
				row.WriteString(a_t)
				row.WriteString(pp.tableColSep)
				a_t = ""
			}

		case *qxml.EndElement:
			switch e.Name() {
			case "a:tr":
				if row.Len() > 0 {
					texts.WriteString(row.String())
					texts.WriteString(pp.tableRowSep)
					row.Reset()
					a_t = ""
				}
			case "w:tbl":
				break NEXT
			}
		}
	}

	return texts
}

// extractImage extracts text content from image by the ocr interface.
//
// Parameters:
//   - i: the index of the slide
//   - rId: the reference id of the image.
//
// Returns:
//   - *strings.Builder: the formatted text of the extracted content.
//   - error: any error that occurred during the extraction process.
func (pp *PPTx) extractImage(i int, rId string) (*strings.Builder, error) {
	if pp.ocr == nil {
		return nil, errors.New("no ocr client")
	}
	if rId == "" {
		return nil, fmt.Errorf("empty rID at index %d", i)
	}

	slideRels, ok := pp.slideRelsMap[i]
	if !ok {
		return nil, fmt.Errorf("slide rels at index %d", i)
	}

	fname, ok := slideRels[rId]
	if !ok {
		return nil, fmt.Errorf("no slide rel with rID %s", rId)
	}

	f, ok := pp.imagesFiles[fname]
	if !ok {
		return nil, fmt.Errorf("no image file with name %s", fname)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	text, err := pp.ocr.Run(rc)
	if err != nil {
		return nil, err
	}

	var (
		fmtTexts = new(strings.Builder)
		lineSep  = "\n"
	)

	if pp.drawingsNoFmt {
		fmtTexts.WriteString(text)
		fmtTexts.WriteString(lineSep)
		return fmtTexts, nil
	}

	var (
		newText, maxLineLen = MaxLineLenWithPrefix(text, []byte(" "))
		halfLine            = bytes.Repeat([]byte("─"), max((maxLineLen-5)/2, 0))
	)

	fmtTexts.WriteString("┌")
	fmtTexts.Write(halfLine)
	fmtTexts.WriteString("image")
	fmtTexts.Write(halfLine)
	fmtTexts.WriteString("┐")
	fmtTexts.WriteString(lineSep)

	fmtTexts.WriteString(newText)
	fmtTexts.WriteString(lineSep)

	fmtTexts.WriteString("└")
	fmtTexts.Write(halfLine)
	fmtTexts.WriteString("─────")
	fmtTexts.Write(halfLine)
	fmtTexts.WriteString("┘")
	fmtTexts.WriteString(lineSep)

	return fmtTexts, nil
}

// extractChart extracts the chart text from the pptx file for a given slide index and relationship ID.
//
// Parameters:
//   - i: the index of the slide
//   - rId: the relationship ID of the chart
//
// Returns:
//   - *strings.Builder: the extracted chart text as a strings.Builder
//   - error: any error that occurred during the extraction
func (pp *PPTx) extractChart(i int, rId string) (*strings.Builder, error) {
	if rId == "" {
		return nil, fmt.Errorf("no rID at index %d", i)
	}

	slideRels, ok := pp.slideRelsMap[i]
	if !ok {
		return nil, fmt.Errorf("no slide rels at index %d", i)
	}

	fname, ok := slideRels[rId]
	if !ok {
		return nil, fmt.Errorf("no slide found in slide rels with rID %s", rId)
	}

	f, ok := pp.chartsFiles[fname]
	if !ok {
		return nil, fmt.Errorf("no chat file found with filename %s", fname)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	var (
		fmtTexts   = new(strings.Builder)
		texts      = new(strings.Builder)
		line       = new(strings.Builder)
		c_v        = ""
		lineSep    = "\n"
		space      = " "
		maxLineLen = 0
	)

	r := qxml.NewReader(rc)
	valRegex := regexp.MustCompile(`(?i)c:.?val`)

NEXT:
	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.EndElement:
			if e.Name() == "c:plotArea" {
				if pp.drawingsNoFmt {
					fmtTexts.WriteString(texts.String())
					fmtTexts.WriteString(lineSep)
					return fmtTexts, nil
				}

				halfLine := bytes.Repeat([]byte("─"), max((maxLineLen-5)/2, 0))
				fmtTexts.WriteString("┌")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("chart")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("┐")
				fmtTexts.WriteString(lineSep)

				fmtTexts.WriteString(texts.String())

				fmtTexts.WriteString("└")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("─────")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("┘")
				fmtTexts.WriteString(lineSep)

				texts.Reset()
				break NEXT
			}
		case *qxml.StartElement:
			switch e.Name() {
			case "c:ser":
			INNER_NEXT:
				for r.Next() {
					switch e := r.Element().(type) {
					case *qxml.EndElement:
						if e.Name() == "c:ser" {
							break INNER_NEXT
						}
					case *qxml.StartElement:
						name := e.Name()
						switch {
						case name == "c:tx":
							if FindNameIterTo(r, "c:v", "c:tx") {
								r.AssignNext(&c_v)
								if !r.Next() {
									break NEXT
								}

								line.WriteString(" [")
								line.WriteString(c_v)
								line.WriteString("]")
								c_v = ""

								if line.Len() > 0 {
									texts.WriteString(line.String())
									texts.WriteString(lineSep)
									if line.Len() > maxLineLen {
										maxLineLen = line.Len()
									}
									line.Reset()
								}
							}
						case name == "c:cat":
							for FindNameIterTo(r, "c:v", "c:cat") {
								r.AssignNext(&c_v)
								if !r.Next() {
									break NEXT
								}
								line.WriteString(c_v)
								line.WriteString(space)
								c_v = ""
							}
							if line.Len() > 0 {
								texts.WriteString(space)
								texts.WriteString(line.String())
								texts.WriteString(lineSep)
								if line.Len() > maxLineLen {
									maxLineLen = line.Len()
								}
								line.Reset()
							}
						case valRegex.MatchString(name):
							for MatchNameIterTo(r, "c:v", `(?i)c:.?val`) {
								r.AssignNext(&c_v)
								if !r.Next() {
									break NEXT
								}
								line.WriteString(c_v)
								line.WriteString(space)
								c_v = ""
							}
							if line.Len() > 0 {
								texts.WriteString(space)
								texts.WriteString(line.String())
								texts.WriteString(lineSep)
								if line.Len() > maxLineLen {
									maxLineLen = line.Len()
								}
								line.Reset()
							}
						}
					}
				}
			}
		}
	}
	return fmtTexts, nil
}

// extractDiagram extracts the diagram text from the pptx file for a given slide index and relationship ID.
//
// Parameters:
//   - i: the index of the slide
//   - rId: the relationship ID of the diagram
//
// Returns:
//   - *strings.Builder: the extracted diagram text as a strings.Builder
//   - error: any error that occurred during the extraction
func (pp *PPTx) extractDiagram(i int, rId string) (*strings.Builder, error) {
	if rId == "" {
		return nil, fmt.Errorf("no rID found at index %d", i)
	}

	slideRels, ok := pp.slideRelsMap[i]
	if !ok {
		return nil, fmt.Errorf("no slide rels found at index %d", i)
	}

	fname, ok := slideRels[rId]
	if !ok {
		return nil, fmt.Errorf("no slide found in slidRels with rID %s", rId)
	}

	f, ok := pp.diagramsFiles[fname]
	if !ok {
		return nil, fmt.Errorf("no diagram file found with filename %s", fname)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var (
		fmtTexts   = new(strings.Builder)
		texts      = new(strings.Builder)
		line       = new(strings.Builder)
		c_v        = ""
		lineSep    = "\n"
		space      = " "
		maxLineLen = 0
	)

	r := qxml.NewReader(rc)

NEXT:
	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.EndElement:
			if e.Name() == "dgm:ptLst" {
				if pp.drawingsNoFmt {
					fmtTexts.WriteString(texts.String())
					fmtTexts.WriteString(lineSep)
					return fmtTexts, nil
				}

				halfLine := bytes.Repeat([]byte("─"), max((maxLineLen-7)/2, 0))
				fmtTexts.WriteString("┌")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("diagram")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("┐")
				fmtTexts.WriteString(lineSep)

				fmtTexts.WriteString(texts.String())

				fmtTexts.WriteString("└")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("───────")
				fmtTexts.Write(halfLine)
				fmtTexts.WriteString("┘")
				fmtTexts.WriteString(lineSep)

				texts.Reset()
				break NEXT
			}
		case *qxml.StartElement:
			switch e.Name() {
			case "a:p":
				for FindNameIterTo(r, "a:t", "a:p") {
					r.AssignNext(&c_v)
					if !r.Next() {
						break NEXT
					}
					line.WriteString(c_v)
					line.WriteString(space)
					c_v = ""
				}
				if line.Len() > 0 {
					texts.WriteString(space)
					texts.WriteString(line.String())
					texts.WriteString(lineSep)
					if line.Len() > maxLineLen {
						maxLineLen = line.Len()
					}
					line.Reset()
				}
			}
		}
	}

	return fmtTexts, nil
}
