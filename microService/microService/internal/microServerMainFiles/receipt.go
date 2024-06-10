package microServerMainFiles

import (
	"bytes"
	"fmt"
	"log"

	"github.com/signintech/gopdf"
)

type ReceiptData struct {
	ProjectName       string
	TransactionNumber string
	Date              string
	Time              string
	CustomerName      string
	PaymentMethod     string
	Items             []ReceiptItem
	GrandTotal        string
}

type ReceiptItem struct {
	Name     string
	Price    string
	Quantity int
	Total    string
}

func GenerateReceiptPDF(transaction *Transaction, customerName string) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 210, H: 297}}) // A4 size in mm
	pdf.AddPage()

	// Add Arial font
	fontPath := "internal/microServerMainFiles/arial.ttf"
	err := pdf.AddTTFFont("arial", fontPath)
	if err != nil {
		log.Printf("Error adding font: %v", err)
		return nil, err
	}

	// Set font size to smaller
	err = pdf.SetFont("arial", "", 8)
	if err != nil {
		log.Printf("Error setting font: %v", err)
		return nil, err
	}

	// Define margin and line height
	marginLeft := 10.0
	lineHeight := 10.0
	textWidth := 190.0

	// Function to draw text
	drawText := func(text string, y float64, x float64) {
		pdf.SetX(x)
		pdf.SetY(y)
		pdf.CellWithOption(&gopdf.Rect{W: textWidth, H: lineHeight}, text, gopdf.CellOption{Align: gopdf.Left})
	}

	// Header
	currentY := 20.0
	drawText("TIN: 123456789", currentY, marginLeft)
	currentY += lineHeight
	drawText("Welcome to our shop", currentY, marginLeft)

	// Transaction details
	currentY += 2 * lineHeight
	drawText(fmt.Sprintf("Project: %s", "Book Shop"), currentY, marginLeft)
	currentY += lineHeight
	drawText(fmt.Sprintf("Transaction #: %s", transaction.ID.Hex()), currentY, marginLeft)
	currentY += lineHeight
	drawText(fmt.Sprintf("Date: %s", transaction.CreatedAt.Format("2006-01-02")), currentY, marginLeft)
	currentY += lineHeight
	drawText(fmt.Sprintf("Time: %s", transaction.CreatedAt.Format("15:04:05")), currentY, marginLeft)
	currentY += lineHeight
	drawText(fmt.Sprintf("Customer: %s", customerName), currentY, marginLeft)
	currentY += lineHeight
	drawText(fmt.Sprintf("Payment Method: %s", "Credit Card"), currentY, marginLeft)

	// Table header
	currentY += 2 * lineHeight
	drawText("Item", currentY, marginLeft)
	drawText("Price", currentY, marginLeft+60)
	drawText("Quantity", currentY, marginLeft+110)
	drawText("Total", currentY, marginLeft+160)

	for _, item := range transaction.Items {
		total := item.Price * float64(item.Quantity)
		currentY += lineHeight
		drawText(item.ProductID, currentY, marginLeft)
		drawText(fmt.Sprintf("$%.2f", item.Price), currentY, marginLeft+60)
		drawText(fmt.Sprintf("%d", item.Quantity), currentY, marginLeft+110)
		drawText(fmt.Sprintf("$%.2f", total), currentY, marginLeft+160)
	}

	// Grand total
	currentY += 2 * lineHeight
	drawText(fmt.Sprintf("Grand Total: $%.2f", transaction.TotalAmount), currentY, marginLeft)

	// Footer
	currentY += 2 * lineHeight
	drawText("THANK YOU", currentY, marginLeft)
	currentY += lineHeight
	drawText("COME BACK AGAIN", currentY, marginLeft)
	currentY += lineHeight

	var pdfBuf bytes.Buffer
	err = pdf.Write(&pdfBuf)
	if err != nil {
		log.Printf("Error writing PDF: %v", err)
		return nil, err
	}

	return pdfBuf.Bytes(), nil
}
