package invoice

import (
	"fmt"
	"time"
)

func GenerateInvoiceNo(date time.Time, sequence int) string {
	return fmt.Sprintf("INV-%s-%04d", date.Format("20060102"), sequence)
}
