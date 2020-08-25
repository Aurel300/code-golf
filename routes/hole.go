package routes

import (
	"net/http"

	"github.com/code-golf/code-golf/hole"
	"github.com/code-golf/code-golf/lang"
	"github.com/code-golf/code-golf/session"
)

// Hole serves GET /{hole}
func Hole(w http.ResponseWriter, r *http.Request) {
	data := struct {
		HideDetails  bool
		Hole         hole.Hole
		Langs        []lang.Lang
		ScoringModes []string
		Solutions    []map[string]string
	}{
		Langs:        lang.List,
		ScoringModes: []string{"Chars"},
		Solutions:    []map[string]string{{}, {}},
	}

	var ok bool
	if data.Hole, ok = hole.ByID[param(r, "hole")]; !ok {
		NotFound(w, r)
		return
	}

	if c, _ := r.Cookie("hide-details"); c != nil {
		data.HideDetails = true
	}

	if golfer := session.Golfer(r); golfer != nil {
		// Fetch all the code per lang.
		condition := ""
		if !session.Beta(r) {
			condition = " AND scoring = 'chars'"
		}

		rows, err := session.Database(r).Query(
			`SELECT code, lang, scoring
			   FROM solutions
			   JOIN code ON code_id = id
			  WHERE hole = $1 AND user_id = $2`+condition,
			data.Hole.ID, golfer.ID,
		)
		if err != nil {
			panic(err)
		}

		defer rows.Close()

		for rows.Next() {
			var code, lang, scoring string

			if err := rows.Scan(&code, &lang, &scoring); err != nil {
				panic(err)
			}

			slot := 0
			if scoring == "bytes" {
				slot = 1
			}

			data.Solutions[slot][lang] = code
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	}

	if session.Beta(r) {
		data.ScoringModes = append(data.ScoringModes, "Bytes")
	}

	render(w, r, "hole", data.Hole.Name, data)
}
