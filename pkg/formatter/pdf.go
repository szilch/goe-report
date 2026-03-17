package formatter

import (
	"bytes"
	_ "embed"
	"echarge-report/pkg/models"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jung-kurt/gofpdf"
)

//go:embed logo.png
var logoBytes []byte

type PDFFormatter struct {
	filename string
}

func NewPDFFormatter(filename string) *PDFFormatter {
	return &PDFFormatter{
		filename: filename,
	}
}

func (f *PDFFormatter) Format(data models.ReportData) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	title := "Ladebericht - Wallbox"

	pdf.Cell(40, 10, tr(title))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	yAbrechnung := pdf.GetY()
	pdf.Cell(40, 8, tr("Abrechnungsdaten"))

	// Add embedded logo, right-aligned
	if len(logoBytes) > 0 {
		logoReader := bytes.NewReader(logoBytes)
		// Register the image from the reader
		pdf.RegisterImageOptionsReader("logo", gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, logoReader)
		// Place the registered image
		// A4 is 210mm wide. 10mm margin. 210 - 10 - 40 (width) = 160
		pdf.ImageOptions("logo", 160, yAbrechnung, 40, 0, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)

	licPlate := data.LicensePlate
	if licPlate == "" {
		licPlate = "Keines hinterlegt"
	}
	pdf.Cell(40, 6, tr(fmt.Sprintf("Kfz-Kennzeichen: %s", licPlate)))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Kilometerstand (%s): %s", time.Now().Format("02.01.2006"), FormatMileage(data.Mileage))))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Kilometerstand (%s): %s", data.EndDate.Format("02.01.2006"), FormatMileage(data.MileageAtEnd))))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Zeitraum: %s - %s", data.StartDate.Format("02.01.2006"), data.EndDate.Format("02.01.2006"))))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Preis/kWh: %s", FormatKWhPrice(data.KwhPrice))))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 8, tr("Start"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 8, tr("Ende"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, tr("Dauer"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(45, 8, tr("Lademenge (kWh)"), "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, tr("Preis"), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 11)

	if data.TotalSessions == 0 {
		pdf.CellFormat(190, 10, tr("Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden."), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	} else {
		for _, session := range data.Sessions {
			pdf.CellFormat(40, 8, tr(session.StartDate.Format("02.01.2006 15:04")), "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 8, tr(session.EndDate.Format("02.01.2006 15:04")), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 8, tr(session.Duration), "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 8, tr(fmt.Sprintf("%.2f kWh", session.Energy)), "1", 0, "R", false, 0, "")
			pdf.CellFormat(40, 8, tr(FormatPrice(session.Price)), "1", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}

		pdf.SetFont("Arial", "B", 12)
		summaryText := fmt.Sprintf("Summe (%d Ladevorgänge)", data.TotalSessions)
		pdf.CellFormat(105, 8, tr(summaryText), "1", 0, "R", false, 0, "")
		pdf.CellFormat(45, 8, tr(fmt.Sprintf("%.2f kWh", data.TotalEnergy)), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, tr(FormatPrice(data.TotalPrice)), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	err := pdf.OutputFileAndClose(f.filename)
	if err != nil {
		return fmt.Errorf("error saving PDF (%s): %w", f.filename, err)
	}

	color.Blue(tr("PDF-Bericht erfolgreich erstellt unter: %s"), f.filename)
	return nil
}
