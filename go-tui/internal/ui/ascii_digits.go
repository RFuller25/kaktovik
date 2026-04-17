package ui

// artWidth and artHeight are the fixed dimensions of every digit glyph.
const artWidth = 9
const artHeight = 7

// digitArt holds 9-wide × 7-tall ASCII art for each Kaktovik numeral 0–19.
//
// Visual grammar
//
//	Tally bars (rows 0–3): 0–4 heavy horizontal lines (━━━━━━━━━)
//	Base shapes (rows 4–6):
//	  none   → digits 0–4
//	  right elbow (━━━━━━━━┓ / ┃) → digit 5, multiples-of-5 group
//	  left  elbow (┏━━━━━━━━ / ┃) → digit 10 group
//	  cross        (┏━━━━━━━┓ / ┃ ┃) → digit 15 group
//	  circle (digit 0 only)
var digitArt = [20][artHeight]string{
	// 0 — circle
	{
		"         ",
		"  ╭───╮  ",
		"  │   │  ",
		"  │   │  ",
		"  ╰───╯  ",
		"         ",
		"         ",
	},
	// 1 — one tally
	{
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"         ",
		"         ",
		"         ",
	},
	// 2 — two tallies
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"         ",
		"         ",
	},
	// 3 — three tallies
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"         ",
	},
	// 4 — four tallies
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
	},
	// 5 — right elbow only
	{
		"         ",
		"         ",
		"         ",
		"         ",
		"━━━━━━━━┓",
		"        ┃",
		"        ┃",
	},
	// 6 — one tally + right elbow
	{
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"━━━━━━━━┓",
		"        ┃",
		"        ┃",
	},
	// 7 — two tallies + right elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"━━━━━━━━┓",
		"        ┃",
		"        ┃",
	},
	// 8 — three tallies + right elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"━━━━━━━━┓",
		"        ┃",
		"        ┃",
	},
	// 9 — four tallies + right elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━┓",
		"        ┃",
		"        ┃",
	},
	// 10 — left elbow only
	{
		"         ",
		"         ",
		"         ",
		"         ",
		"┏━━━━━━━━",
		"┃        ",
		"┃        ",
	},
	// 11 — one tally + left elbow
	{
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"┏━━━━━━━━",
		"┃        ",
		"┃        ",
	},
	// 12 — two tallies + left elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"┏━━━━━━━━",
		"┃        ",
		"┃        ",
	},
	// 13 — three tallies + left elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"┏━━━━━━━━",
		"┃        ",
		"┃        ",
	},
	// 14 — four tallies + left elbow
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"┏━━━━━━━━",
		"┃        ",
		"┃        ",
	},
	// 15 — cross (both elbows)
	{
		"         ",
		"         ",
		"         ",
		"         ",
		"┏━━━━━━━┓",
		"┃       ┃",
		"┃       ┃",
	},
	// 16 — one tally + cross
	{
		"━━━━━━━━━",
		"         ",
		"         ",
		"         ",
		"┏━━━━━━━┓",
		"┃       ┃",
		"┃       ┃",
	},
	// 17 — two tallies + cross
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"         ",
		"┏━━━━━━━┓",
		"┃       ┃",
		"┃       ┃",
	},
	// 18 — three tallies + cross
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"         ",
		"┏━━━━━━━┓",
		"┃       ┃",
		"┃       ┃",
	},
	// 19 — four tallies + cross
	{
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"━━━━━━━━━",
		"┏━━━━━━━┓",
		"┃       ┃",
		"┃       ┃",
	},
}

// digitLines returns the artHeight ASCII rows for Kaktovik numeral n (0–19).
func digitLines(n int) [artHeight]string {
	if n < 0 || n > 19 {
		return digitArt[0]
	}
	return digitArt[n]
}
