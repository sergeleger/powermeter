package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/urfave/cli"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// GraphCommand collects power usage information from the command line.
var GraphCommand = cli.Command{
	Name:      "graph",
	Usage:     "Creates a plot for the specified parameter",
	ArgsUsage: "dbFile year [ month [ day ] ]",
	Action:    graphAction,
}

func graphAction(ctx *cli.Context) error {
	args := ctx.Args()

	return createMonthly(args[0], args[1])

	// switch ctx.NArg() {
	// case 2:
	// 	createMonthly(ctx.Args().First(), ctx.Args().Get(1))
	// case 3:
	// 	createDaily(ctx.Args().First(), ctx.Args().Get(1), ctx.Args().Get(2))
	// case 4:
	//     createHourly(ctx.Args

	// default:
	// 	return errors.New("error: not enough arguments")
	// }

	return nil
}

func createMonthly(dbFile, year string) error {
	db, err := openDatabase(dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	st, err := db.Prepare(`select  Time, Usage from power_by_day where MeterID = 18011759 order by 2`)
	if err != nil {
		return err
	}
	defer st.Close()

	r, err := st.Query()
	if err != nil {
		return err
	}

	var tZ string
	var usage float64
	var t time.Time

	var values plotter.Values
	var labels []string
	for r.Next() {
		r.Scan(&tZ, &usage)
		t, _ = time.Parse("2006-01-02 15:04:05.999999999Z07:00", tZ)

		values = append(values, usage)
		labels = append(labels, t.Format("02"))
		fmt.Println(t)
	}

	for j := len(values) - 1; j > 0; j-- {
		values[j] = values[j] - values[j-1]
	}

	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = "Daily Power Usage"

	w := vg.Points(5)

	barsA, err := plotter.NewBarChart(values[1:], w)
	if err != nil {
		panic(err)
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = color.RGBA{R: 255, A: 255}
	//barsA.Offset = -w

	p.Add(barsA)
	p.NominalX(labels[1:]...)

	if err := p.Save(20*vg.Inch, 3*vg.Inch, "barchart.png"); err != nil {
		panic(err)
	}

	return nil
}
