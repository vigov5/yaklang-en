package scanfpcmd

import (
	"context"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bruteutils"
	"os"
)

var BruteUtil = cli.Command{
	Name: "brute",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "target,t",
		},
		cli.StringFlag{
			Name:  "username,u",
			Value: "dataex/dicts/user.txt",
		},
		cli.StringFlag{
			Name:  "password,p",
			Value: "dataex/dicts/3389.txt",
		},
		cli.IntFlag{
			Name:  "min-delay",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "max-delay",
			Value: 2,
		},
		cli.IntFlag{
			Name:  "target-concurrent",
			Value: 200,
		},
		cli.StringFlag{
			Name: "type,x",
		},
		cli.BoolFlag{
			Name:  "ok-to-stop",
			Usage: "If a target finds a successful result, stop blasting",
		},
		cli.IntFlag{
			Name:  "finished-to-end",
			Usage: "Exploded results if displayed multiple times'Finished' on this target. This option controls the threshold.",
			Value: 10,
		},
		cli.StringFlag{
			Name:  "divider",
			Usage: "User (username), password (password), input separator, the default is (,)",
			Value: ",",
		},
	},

	Action: func(c *cli.Context) error {
		bruteFunc, err := bruteutils.GetBruteFuncByType(c.String("type"))
		if err != nil {
			return err
		}

		bruter, err := bruteutils.NewMultiTargetBruteUtil(
			c.Int("target-concurrent"), c.Int("min-delay"), c.Int("max-delay"),
			bruteFunc,
		)
		if err != nil {
			return err
		}

		bruter.OkToStop = c.Bool("ok-to-stop")
		bruter.FinishingThreshold = c.Int("finished-to-end")

		var succeedResult []*bruteutils.BruteItemResult

		userList := bruteutils.FileOrMutateTemplate(c.String("username"), c.String("divider"))
		err = bruter.StreamBruteContext(
			context.Background(), c.String("type"),
			bruteutils.FileOrMutateTemplateForStrings(c.String("divider"), utils.ParseStringToHosts(c.String("target"))...),
			userList,
			bruteutils.FileOrMutateTemplate(c.String("password"), c.String("divider")),
			func(b *bruteutils.BruteItemResult) {
				if b.Ok {
					succeedResult = append(succeedResult, b)
					log.Infof("Success for target: %v user: %v pass: %s", b.Target, b.Username, b.Password)
				} else {
					log.Warningf("failed for target: %v user: %v pass: %s", b.Target, b.Username, b.Password)
				}
			},
		)
		if err != nil {
			return err
		}

		log.Info("------------------------------------------------")
		log.Info("------------------------------------------------")
		log.Info("------------------------------------------------")

		if len(succeedResult) <= 0 {
			log.Info("does not blast to Available results")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Service type", "target", "User name", "Password",
		})
		for _, i := range succeedResult {
			if i.OnlyNeedPassword {
				table.Append([]string{
					c.String("type"),
					i.Target,
					"",
					i.Password,
				})
			} else {
				table.Append([]string{
					c.String("type"),
					i.Target,
					i.Username,
					i.Password,
				})
			}
		}
		table.Render()

		return nil
	},
}
