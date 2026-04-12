package cmd

import (
	"context"

	"myjob/internal/bootstrap"

	"github.com/gogf/gf/v2/os/gcmd"
)

var Main = &gcmd.Command{
	Name:  "main",
	Brief: "run myjob admin backend",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		app, err := bootstrap.NewApplicationFromEnv()
		if err != nil {
			return err
		}
		defer app.Close()
		app.Server().Run()
		return nil
	},
}
