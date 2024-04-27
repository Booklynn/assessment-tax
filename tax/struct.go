package tax

type Allowances struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type TaxInfo struct {
	TotalIncome float64      `json:"totalIncome"`
	WHT         float64      `json:"wht"`
	Allowances  []Allowances `json:"allowances"`
}

type TaxPayable struct {
	Tax float64 `json:"tax"`
}
