package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
)

func main() {

	TASK_CREATE_SUBCMD := cli.Command{
		Name:   "create",
		Action: CreateTaskCmd,
		Flags: []cli.Flag{
			InteractiveFlag(),
			TaskAttributeFlag(),
			JsonTaskAttributeFlag(),
		},
	}

	TASK_SEARCH_SUBCMD := cli.Command{
		Name:   "search",
		Action: SearchByMeta,
	}

	TASK_BACKLINKS_SUBCMD := cli.Command{
		Name:   "backlinks",
		Action: GenerateBacklinks,
	}

	TASK_CMD := cli.Command{
		Name: "task",
		Flags: []cli.Flag{
			PathFlag(),
		},
		Subcommands: []*cli.Command{
			&TASK_CREATE_SUBCMD,
			&TASK_SEARCH_SUBCMD,
			&TASK_BACKLINKS_SUBCMD,
		},
	}

	app := &cli.App{
		Commands: []*cli.Command{
			&TASK_CMD,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func SearchByMeta(cCtx *cli.Context) error {
	root := cCtx.String("path")
	r := FileTaskRepo{Root: root}
	meta, err := MakeYamlWithEditor()
	if err != nil {
		return err
	}

	tasks, err := r.FindTasksByMeta(meta)
	if err != nil {
		return err
	}

	for i, t := range tasks {
		fmt.Println(i, ": ", t)
	}

	return nil
}

func GenerateBacklinks(cCtx *cli.Context) error {
	root := cCtx.String("path")
	file := cCtx.Args().Get(0)
	r := FileTaskRepo{Root: root}

	p := root + "/" + file + ".md"
	t, err := r.LoadTask(p)
	if err != nil {
		fmt.Println("Error loading task: ", err)
		return err
	}

	backlinks := ExtractLinks(&t.Contents)

	for _, b := range backlinks {
		p := root + b.Ref
		bt, err := r.LoadTask(p)
		fmt.Println("Loading task: ", p)
		fmt.Println("Task Path: ", bt.Path)
		if err != nil {
			fmt.Println("Error loading task: ", b)
		}

		cbl, ok := bt.Meta["backlinks"].([]string)
		if ok {
			cbl = append(cbl, "."+t.Path)
		} else {
			cbl = []string{"." + t.Path}
		}

		bt.Meta["backlinks"] = cbl

		if err := r.SaveTask(bt); err != nil {
			fmt.Println("Error saving task: ", bt)
			fmt.Println("Error: ", err)
		}
	}

	return nil
}

func CreateTaskCmd(cCtx *cli.Context) error {
	if cCtx.Bool("interactive") {
		return CreateTaskEditor(cCtx)
	}
	return CreateTaskFile(cCtx)
}

func CreateTaskPath(r *FileTaskRepo, title string) (string, error) {
	tt := title

	for i := 0; ; i++ {
		if _, err := r.SearchByTitle(tt); err.Error() == "Task not found" {
			break
		}

		tt = title + "." + string(i+1)
	}

	tp := strings.ReplaceAll(tt, " ", "_") + ".md"

	return tp, nil
}

func CreateTaskEditor(cCtx *cli.Context) error {
	root := cCtx.String("path")
	r := FileTaskRepo{Root: root}

	title := cCtx.Args().Get(0)
	if title == "" {
		return fmt.Errorf("Title is required")
	}

	p, err := CreateTaskPath(&r, title)
	if err != nil {
		return err
	}

	t := &Task{Path: p, Meta: Dict{"title": title}}

	if err := r.EditTask(t); err != nil {
		return err
	}

	if err := r.SaveTask(t); err != nil {
		return err
	}

	return nil
}

func CreateTaskFile(cCtx *cli.Context) error {
	root := cCtx.String("path")
	r := FileTaskRepo{Root: root}

	title := cCtx.Args().Get(0)
	if title == "" {
		return fmt.Errorf("Title is required")
	}

	tp := strings.ReplaceAll(title, " ", "_") + ".md"

	task, err := r.NewTask(Dict{"title": title}, "", tp)
	if err != nil {
		return err
	}

	var attr Dict

	if a := cCtx.String("attribute"); a != "" {
		s := strings.Split(a, "=")

		if len(s) != 2 {
			log.Fatal("Invalid attribute")
		}

		attr = Dict{s[0]: s[1]}
	}

	if a := cCtx.String("json-attr"); a != "" && attr == nil {
		s := strings.Split(a, "=")
		if len(s) != 2 {
			log.Fatal("Invalid attribute")
		}

		if err := json.Unmarshal([]byte(s[1]), &attr); err != nil {
			return err
		}
	}

	fmt.Println(attr)

	for k, v := range attr {
		if task.Meta[k] == nil {
			task.Meta[k] = v
		}
	}

	if err := r.SaveTask(task); err != nil {
		return err
	}

	return nil
}

func TaskAttributeFlag() cli.Flag {
	return &cli.StringFlag{
		Name:    "attribute",
		Aliases: []string{"a"},
		Usage:   "Attribute `ATTR` to set",
	}
}

func JsonTaskAttributeFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "json-attr",
		Usage: "Attribute `ATTR` to set",
	}
}

func InteractiveFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "interactive",
		Aliases: []string{"i"},
	}
}

func PathFlag() cli.Flag {
	return &cli.StringFlag{
		Name:    "path",
		Value:   "./",
		Aliases: []string{"p"},
	}
}
