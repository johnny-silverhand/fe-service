package model

type ColumnOrder struct {
	Column string
	Type   string
}

func GetOrder(s string) ColumnOrder {

	order := ColumnOrder{Type: "ASC", Column: "Id"}

	if len(s) > 0 {
		var descending = false
		if s[0:1] == "-" {
			descending = true
			s = s[1:]
		}

		if descending == true {
			order.Type = "DESC"
		}
	}

	return order
}
