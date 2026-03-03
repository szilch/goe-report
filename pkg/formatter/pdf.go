package formatter

import (
	"fmt"

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

	// Header for table (spanning total width of 190mm for A4 with 10mm margins)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(65, 8, tr("Datum"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, tr("Dauer"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(65, 8, tr("Lademenge (kWh)"), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 11)

	if data.TotalSessions == 0 {
		pdf.CellFormat(190, 10, tr("Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden."), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	} else {
		// Rows
		for _, session := range data.Sessions {
			pdf.CellFormat(65, 8, tr(session.Date), "1", 0, "C", false, 0, "")
			pdf.CellFormat(60, 8, tr(session.Duration), "1", 0, "C", false, 0, "")
			pdf.CellFormat(65, 8, fmt.Sprintf("%.2f", session.Energy), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}

		// Summary Row
		pdf.SetFont("Arial", "B", 12)
		summaryText := fmt.Sprintf("Summe (%d Ladevorgänge)", data.TotalSessions)
		pdf.CellFormat(125, 8, tr(summaryText), "1", 0, "R", false, 0, "")
		pdf.CellFormat(65, 8, fmt.Sprintf("%.2f", data.TotalEnergy), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	err := pdf.OutputFileAndClose(f.filename)
	if err != nil {
		return fmt.Errorf("fehler beim Speichern der PDF (%s): %w", f.filename, err)
	}

	fmt.Printf(tr("PDF-Bericht erfolgreich erstellt unter: %s\n"), f.filename)
	return nil
}
