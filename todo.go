package main

import (
	"fmt"
	// "io/ioutil"
	// "os"
	// "gopkg.in/yaml.v3"
	"regexp"
)

type (
	Dict map[string]interface{}

	Task struct {
		Path     string
		Meta     Dict
		Contents string
	}

	TaskRepo interface {
		NewTask(meta Dict, contents string, path string) (*Task, error)
		SaveTask(task *Task) error
		LoadTask(path string) (*Task, error)
		SearchByMeta(path string, meta Dict) ([]*Task, error)
	}

	ToDict interface {
		AtoDict() Dict
	}

	Link struct {
		Title string
		Ref   string
	}
)

func ExtractLinks(s *string) []Link {
	links := []Link{}
	fmt.Println("Extracting links from: ", "\n\n```", *s, "\n```\n")

	p := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	m := p.FindAllStringSubmatch(*s, -1)

	fmt.Println("Found links: ", m[0][2])

	for _, v := range m {
		links = append(links, Link{Title: v[1], Ref: v[2]})
	}

	return links
}

func (d *Dict) DeepContains(a *Dict) bool {
	for k, v := range *a {
		// if (*d)[k] is Dict
		dD, dOk := (*d)[k].(Dict)
		aD, aOk := v.(Dict)
		if dOk && aOk {
			if !dD.DeepContains(&aD) {
				return false
			}
		} else {
			if (*d)[k] != v {
				return false
			}
		}
	}
	return true
}
