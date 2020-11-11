package sqlstore

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/mattermost/gorp"
	"im/mlog"
)

// Converts a list of strings into a list of query parameters and a named parameter map that can
// be used as part of a SQL query.
func MapStringsToQueryParams(list []string, paramPrefix string) (string, map[string]interface{}) {
	keys := bytes.Buffer{}
	params := make(map[string]interface{}, len(list))
	for i, entry := range list {
		if keys.Len() > 0 {
			keys.WriteString(",")
		}

		key := paramPrefix + strconv.Itoa(i)
		keys.WriteString(":" + key)
		params[key] = entry
	}

	return fmt.Sprintf("(%v)", keys.String()), params
}

// Converts a list of strings into a list of query parameters and a named parameter map that can
// be used as part of a SQL query.
func StringsToQueryParams(list []string) (string, []interface{}) {
	keys := bytes.Buffer{}
	params := make([]interface{}, 0)
	for _, entry := range list {
		if keys.Len() > 0 {
			keys.WriteString(",")
		}

		keys.WriteString("?")
		params = append(params, entry)
	}

	return fmt.Sprintf("(%v)", keys.String()), params
}

// finalizeTransaction ensures a transaction is closed after use, rolling back if not already committed.
func finalizeTransaction(transaction *gorp.Transaction) {
	// Rollback returns sql.ErrTxDone if the transaction was already closed.
	if err := transaction.Rollback(); err != nil && err != sql.ErrTxDone {
		mlog.Error("Failed to rollback transaction", mlog.Err(err))
	}
}
