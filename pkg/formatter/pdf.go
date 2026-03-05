package formatter

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jung-kurt/gofpdf"
)

// PDFFormatter outputs the report to a PDF file.
type PDFFormatter struct {
	filename string
}

// NewPDFFormatter creates a new PDFFormatter.
func NewPDFFormatter(filename string) *PDFFormatter {
	return &PDFFormatter{
		filename: filename,
	}
}

// Format generates a PDF containing the ReportData and saves it to the defined filename.
func (f *PDFFormatter) Format(data ReportData) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Create a unicode translator for basic cp1252 which includes common german Umlaute for Arial
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Title
	title := fmt.Sprintf("Ladebericht %s - go-e Wallbox (SN: %s)", data.MonthName, data.SerialNumber)

	pdf.Cell(40, 10, tr(title))
	pdf.Ln(12)

	// Abrechnungsdaten
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 8, tr("Abrechnungsdaten"))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)

	licPlate := data.LicensePlate
	if licPlate == "" {
		licPlate = "Keines hinterlegt"
	}
	pdf.Cell(40, 6, tr(fmt.Sprintf("Kfz-Kennzeichen: %s", licPlate)))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Kilometerstand: %s", data.Mileage)))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Zeitraum: %s", data.MonthName)))
	pdf.Ln(6)
	pdf.Cell(40, 6, tr(fmt.Sprintf("Preis/kWh: %.2f EUR", data.KwhPrice)))
	pdf.Ln(12)

	// Header for table (spanning total width of 190mm for A4 with 10mm margins)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(50, 8, tr("Datum"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(45, 8, tr("Dauer"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(55, 8, tr("Lademenge (kWh)"), "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, tr("Preis (EUR)"), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 11)

	if data.TotalSessions == 0 {
		pdf.CellFormat(190, 10, tr("Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden."), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	} else {
		// Rows
		for _, session := range data.Sessions {
			pdf.CellFormat(50, 8, tr(session.Date), "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 8, tr(session.Duration), "1", 0, "C", false, 0, "")
			pdf.CellFormat(55, 8, tr(fmt.Sprintf("%.2f kWh", session.Energy)), "1", 0, "R", false, 0, "")
			pdf.CellFormat(40, 8, tr(fmt.Sprintf("%.2f €", session.Price)), "1", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}

		// Summary Row
		pdf.SetFont("Arial", "B", 12)
		summaryText := fmt.Sprintf("Summe (%d Ladevorgänge)", data.TotalSessions)
		pdf.CellFormat(95, 8, tr(summaryText), "1", 0, "R", false, 0, "") // 50+45 width
		pdf.CellFormat(55, 8, tr(fmt.Sprintf("%.2f kWh", data.TotalEnergy)), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, tr(fmt.Sprintf("%.2f €", data.TotalPrice)), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	err := pdf.OutputFileAndClose(f.filename)
	if err != nil {
		return fmt.Errorf("fehler beim Speichern der PDF (%s): %w", f.filename, err)
	}

	color.Blue(tr("PDF-Bericht erfolgreich erstellt unter: %s"), f.filename)
	return nil
}
