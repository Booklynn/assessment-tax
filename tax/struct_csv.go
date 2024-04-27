package tax

type TaxCSV struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund"`
}

type TaxResponseCSV struct {
	Taxes []TaxCSV `json:"taxes"`
}
