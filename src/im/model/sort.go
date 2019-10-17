package model
type ColumnOrder struct {
	Column string
	Type   string
}

func GetOrder(s string) ColumnOrder {
	var descending = false
	if s[0:1] == "-" {
		descending = true
		s = s[1:]
	}
	order := ColumnOrder{Type: "ASC", Column: s}
	if descending == true {
		order.Type = "DESC"
	}

	return order
}
