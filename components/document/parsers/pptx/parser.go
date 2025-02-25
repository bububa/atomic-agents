package pptx

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"regexp"
	"strconv"

	"github.com/bububa/atomic-agents/components/document"
)

var (
	re_SLIDE      = regexp.MustCompile(`ppt/slides/slide(\d+)\.xml`)
	re_SLIDE_RELS = regexp.MustCompile(`ppt/slides/_rels/slide(\d+)\.xml\.rels`)
	re_CHARTS     = regexp.MustCompile(`ppt/charts/chart\d+\.xml`)
	re_IMAGES     = regexp.MustCompile(`ppt/media/image\d+\.(?:png|gif|jpg|jpeg)`)
	re_DIAGRAMS   = regexp.MustCompile(`ppt/diagrams/data\d+\.xml`)
)

type Parser struct{}

var _ document.Parser = (*Parser)(nil)

// Parse try to parse a pdf content from a bytes.Reader and write to an io.Writer
func (p *Parser) Parse(ctx context.Context, reader *bytes.Reader, writer io.Writer) error {
	size := reader.Size()
	pptx := NewPPTx()
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return err
	}

	if err = matchZipFile(pptx, zipReader); err != nil {
		pptx.Close()
		return err
	}
	defer pptx.Close()
	content, err := pptx.ExtractTexts()
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(content))
	return err
}

// matchZipFile is a function that matches the files in a zip.Reader to specific categories such as slides, charts,
// images, and diagrams. It populates the relevant maps in the PptxParser struct with the matched files.
//
// Parameters:
//   - pp: a pointer to the PptxParser struct that holds the maps for slideFiles, chartsFiles, imagesFiles,
//     diagramsFiles, and slideRelsMap.
//   - r: a pointer to the zip.Reader struct that contains the files to be matched.
//
// Return:
//   - error: an error if there was a problem parsing the files or populating the maps, otherwise nil.
func matchZipFile(pp *PPTx, r *zip.Reader) error {
	slidesNum := max((len(r.File)-16)/3, 1)
	pp.slideFiles = make(map[int]*zip.File, slidesNum)
	pp.chartsFiles = make(map[string]*zip.File, 4)
	pp.imagesFiles = make(map[string]*zip.File, 4)
	pp.diagramsFiles = make(map[string]*zip.File, 4)
	pp.slideRelsMap = make(map[int]map[string]string, slidesNum)

	for _, file := range r.File {
		switch {
		case re_CHARTS.MatchString(file.Name):
			pp.chartsFiles[file.Name] = file
		case re_IMAGES.MatchString(file.Name):
			pp.imagesFiles[file.Name] = file
		case re_DIAGRAMS.MatchString(file.Name):
			pp.diagramsFiles[file.Name] = file
		default:
			matches := re_SLIDE.FindStringSubmatch(file.Name)
			if len(matches) > 1 {
				i, err := strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				pp.slideFiles[i] = file
				continue
			}

			matches = re_SLIDE_RELS.FindStringSubmatch(file.Name)
			if len(matches) > 1 {
				i, err := strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				relsMap, err := ParseRelsMap(file, "ppt/")
				if err != nil {
					return err
				}
				pp.slideRelsMap[i] = relsMap
				continue
			}
		}
	}

	return nil
}
