package main

import (
	"bufio"
	// "fmt"
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileTaskRepo struct {
	Root string
}

func (F *FileTaskRepo) NewTask(meta Dict, contents string, path string) (*Task, error) {
	return &Task{Path: path, Meta: meta, Contents: contents}, nil
}

func (F *FileTaskRepo) SaveTask(task *Task) error {
	metaYaml, err := yaml.Marshal(task.Meta)
	if err != nil {
		return err
	}

	// taskPath := F.Root + "/" + task.path
	file, err := os.Create(F.Root + task.Path)
	if err != nil {
		return err
	}

	y := "---\n" + string(metaYaml) + "---\n"
	file.WriteString(y + "\n" + task.Contents)
	return nil
}

func (F *FileTaskRepo) SearchByTitle(title string) (*Task, error) {
	path := strings.ReplaceAll(title, " ", "_") + ".md"
	t, err := F.LoadTask(path)
	if err != nil {
		return nil, errors.New("Task not found")
	}
	return t, nil
}

func (F *FileTaskRepo) LoadTask(path string) (*Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines := []string{}
	scanner := bufio.NewScanner(f)
	start := false
	for scanner.Scan() {
		if scanner.Text() == "---" {
			if start {
				break
			}
			start = true
			continue
		}

		lines = append(lines, scanner.Text())
	}

	mStr := strings.Join(lines, "\n")
	meta := Dict{}
	err = yaml.Unmarshal([]byte(mStr), &meta)
	if err != nil {
		return nil, err
	}

	cLines := []string{}
	for scanner.Scan() {
		cLines = append(cLines, scanner.Text())
	}
	contents := strings.Join(cLines, "\n")

	parts := strings.Split(path, "/")
	p := strings.Join(parts[1:], "/")
	return &Task{Path: p, Meta: meta, Contents: contents}, nil
}

func (F *FileTaskRepo) FindTasksByMeta(meta Dict) ([]string, error) {
	todos := []string{}
	todoFilePaths, err := ioutil.ReadDir(F.Root)
	if err != nil {
		return nil, err
	}

	for _, f := range todoFilePaths {
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}
		fp := F.Root + f.Name()
		task, err := F.LoadTask(fp)
		if err != nil {
			return nil, err
		}

		if task.Meta.DeepContains(&meta) {
			todos = append(todos, task.Path)
		}
	}

	return todos, nil
}

func (F *FileTaskRepo) EditTask(task *Task) error {
	tr := FileTaskRepo{Root: os.TempDir()}
	if err := tr.SaveTask(task); err != nil {
		return err
	}

	p := tr.Root + "/" + task.Path

	if err := OpenEditor(p); err != nil {
		return err
	}

	t, err := tr.LoadTask(p)
	if err != nil {
		return err
	}

	task.Contents = t.Contents
	task.Meta = t.Meta

	return nil
}

func OpenEditor(path string) error {
	e := os.Getenv("EDITOR")

	binary, err := exec.LookPath(e)
	if err != nil {
		return err
	}

	s := binary + " " + path
	cmd := exec.Command("sh", "-c", s)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func MakeYamlWithEditor() (Dict, error) {
	tf, err := os.CreateTemp("", "meta.*.yaml")
	if err != nil {
		return nil, err
	}
	OpenEditor(tf.Name())

	meta := Dict{}
	b, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, &meta); err != nil {
		return nil, err
	}

	return meta, nil
}
