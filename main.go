package main

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "importer",
		Usage: "Import edm.su data into MeiliSearch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "http://localhost:7700",
				Usage: "MeiliSearch host",
			},
			&cli.StringFlag{
				Name:  "api-key",
				Value: "",
				Usage: "MeiliSearch API key",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "videos",
				Aliases: []string{"v"},
				Usage:   "Import videos",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "csv",
						Usage:    "CSV file",
						FilePath: "videos.csv",
					},
				},
				Action: func(c *cli.Context) error {
					client := meilisearch.NewClient(meilisearch.ClientConfig{
						Host:   c.String("host"),
						APIKey: c.String("api-key"),
					})

					index := client.Index("videos")
					task, err := index.AddDocumentsCsv([]byte(c.String("csv")))
					if err != nil {
						log.Fatal(err)
					}

					_, err = client.WaitForTask(task)
					if err != nil {
						return err
					}
					fmt.Println("Done")
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
