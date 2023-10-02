package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/meilisearch/meilisearch-go"
	"github.com/urfave/cli/v2"
)

var conn *pgx.Conn

var msClient *meilisearch.Client

type Video struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Date        int    `json:"date"`
	YtThumbnail string `json:"yt_thumbnail"`
	Duration    int    `json:"duration"`
}

func main() {
	app := &cli.App{
		Name:                 "importer",
		EnableBashCompletion: true,
		Usage:                "Import edm.su data into MeiliSearch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "http://localhost:7700",
				Usage: "MeiliSearch host",
				EnvVars: []string{
					"MEILISEARCH_HOST",
				},
			},
			&cli.StringFlag{
				Name:  "api-key",
				Value: "",
				Usage: "MeiliSearch API key",
				EnvVars: []string{
					"MEILISEARCH_API_KEY",
				},
			},
			&cli.StringFlag{
				Name:  "pg-url",
				Value: "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				Usage: "Postgres URL",
				EnvVars: []string{
					"POSTGRES_URL",
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "videos",
				Aliases: []string{"v"},
				Usage:   "Import videos",
				Action:  videosHandler(),
			},
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configure MeiliSearch Indexes",
				Action:  configHandler(),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "videos-index",
						Value: "videos",
						Usage: "MeiliSearch videos index name",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func configHandler() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		initMsClient(c)
		videosIndex := msClient.Index(c.String("videos-index"))
		task, err := videosIndex.UpdateSettings(&meilisearch.Settings{
			SortableAttributes:   []string{"title", "date"},
			FilterableAttributes: []string{"slug", "yt_id"},
			RankingRules: []string{
				"sort",
				"words",
				"typo",
				"proximity",
				"attribute",
				"exactness",
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		_, err = msClient.WaitForTask(task.TaskUID)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	}
}

func videosHandler() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		initMsClient(c)

		var err error
		conn, err = pgx.Connect(context.Background(), c.String("pg-url"))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close(context.Background())

		videos, err := listVideos()
		if err != nil {
			log.Fatal(err)
		}

		index := msClient.Index("videos")
		task, err := index.AddDocuments(videos)
		if err != nil {
			log.Fatal(err)
		}

		_, err = msClient.WaitForTask(task.TaskUID)
		if err != nil {
			return err
		}
		fmt.Println("Done")
		return nil
	}
}

func listVideos() ([]Video, error) {
	rows, _ := conn.Query(
		context.Background(),
		"SELECT"+
			" id,"+
			" title,"+
			" slug,"+
			" extract(epoch from date) as date,"+
			" yt_thumbnail,"+
			" duration"+
			" FROM videos"+
			" WHERE deleted = false",
	)

	var videos []Video

	for rows.Next() {
		var video Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Slug,
			&video.Date,
			&video.YtThumbnail,
			&video.Duration,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}
	return videos, rows.Err()
}

func initMsClient(c *cli.Context) {
	msClient = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   c.String("host"),
		APIKey: c.String("api-key"),
	})
}
