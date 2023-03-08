package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

/********************************************************************************

The purpose of program is to fetch all incomplete tasks.

We start from the beginning of time and go through the list to add items to stack
and then pop them off if they were done. Items that haven't been done and repeated
will be shown as one item.

*********************************************************************************/

var (
	taskre        *regexp.Regexp = regexp.MustCompile(`(?ms)^\s*-\s+\[(.*?)\](.*?)$`)
	taskstrlinkre *regexp.Regexp = regexp.MustCompile(`(\*\[\[.*?\d{4}\|.*?(\.md)?\]\]\*)`)
)

type Task struct {
	Name       string
	StringMD   string
	Done       bool
	OriginFile string
	Created    time.Time
}

type MDFile struct {
	Filename string
	Content  string
	Matter   map[string]string
	Created  time.Time
}

func parseMatter(fileContent []byte) (map[string]string, string) {
	re := regexp.MustCompile(`(?sm)^---\s*$`)
	parts := re.Split(string(fileContent), 3)
	matter, content := parts[1], parts[2]

	frontmatterValues := make(map[string]string)
	yaml.Unmarshal([]byte(matter), &frontmatterValues)

	return frontmatterValues, content
}

func getMDFiles(dirPath string) []*MDFile {

	px(dirPath)

	matches, _ := filepath.Glob(filepath.Join(dirPath, "*.md"))
	mdfiles := make([]*MDFile, 0)
	for _, match := range matches {
		contentBytes, _ := ioutil.ReadFile(match)
		mdfile := &MDFile{}
		mdfile.Filename = match
		mdfile.Matter, mdfile.Content = parseMatter(contentBytes)
		mdfiles = append(mdfiles, mdfile)
		ts, _ := strconv.ParseInt(mdfile.Matter["created"], 10, 64)
		mdfile.Created = time.Unix(ts/1000, 0)
	}
	return mdfiles
}

func filterFileByName(files []*MDFile, filename string) *MDFile {
	for _, mdfile := range files {
		if strings.Contains(mdfile.Filename, filename) {
			return mdfile
		}
	}
	return nil
}

func main() {
	paramFile := flag.String("file", "", "current file")
	paramWrite := flag.Bool("write", false, "write to selected file")
	paramHirearchy := flag.String("hirearchy", "daily.journal", "filter tasks from a hirearchy")
	// flagAutoHirearchy := flag.Bool("auto-hirearchy", false, "auto detect hirearchy")

	flag.Parse()

	if *paramFile == "" {
		log.Fatalln("File path required")

	}

	vaultPath := path.Dir(*paramFile)

	allPendingTasks := map[string]*Task{}

	allMDFiles := getMDFiles(vaultPath)
	sort.SliceStable(allMDFiles, func(i, j int) bool {
		return allMDFiles[i].Created.Before(allMDFiles[j].Created)
	})

	for _, file := range allMDFiles {

		// We want to ignore existing file
		if strings.Contains(*paramFile, file.Filename) {
			continue
		}
		// Include only if hirearchy matches
		if !strings.Contains(file.Filename, *paramHirearchy) {
			continue
		}
		// Ignore if file name has template in it
		if strings.Contains(file.Filename, "template") {
			continue
		}

		taskMatches := taskre.FindAllStringSubmatch(file.Content, -1)

		for _, taskMatch := range taskMatches {
			isDone := taskMatch[1] == "x"
			taskLabel := taskMatch[2]
			taskLabel = taskstrlinkre.ReplaceAllString(taskLabel, "")
			taskLabel = strings.Trim(taskLabel, " \r\n")
			taskLabel = strings.Trim(taskLabel, " \n")
			taskKey := strings.ToLower(taskLabel)

			if taskLabel != "" {
				// Only if the task was not seen before add it
				if _, e := allPendingTasks[taskKey]; !e {
					allPendingTasks[taskKey] = &Task{
						Name:       taskLabel,
						Created:    file.Created,
						Done:       isDone,
						OriginFile: file.Filename,
					}
				} else {
					// If the task was seen but the new one is done
					// then remove it from pending list
					if isDone {
						delete(allPendingTasks, taskKey)
					}
				}
			}
		}

	}

	// Build the pending tasks list, sorted by frontmatter's creation date
	alltasksList := []*Task{}
	alltasksMdStrings := []string{}
	for _, task := range allPendingTasks {
		if !task.Done {
			alltasksList = append(alltasksList, task)
		}
	}
	sort.SliceStable(alltasksList, func(i, j int) bool {
		return alltasksList[i].Created.After(alltasksList[j].Created)
	})

	// Construct the markdown!
	for _, task := range alltasksList {
		destFile := path.Base(task.OriginFile)
		destFile = destFile[:strings.LastIndex(destFile, ".md")]
		taskString := fmt.Sprintf("- [ ] %s *[[%s|%s]]*", task.Name, task.Created.Format("2 Jan 2006"), destFile)
		alltasksMdStrings = append(alltasksMdStrings, taskString)
	}

	// px(alltasks)

	tasksString := strings.Join(alltasksMdStrings, "\n")
	fmt.Println(tasksString)

	if *paramWrite {
		file := filterFileByName(allMDFiles, *paramFile)
		if file != nil {
			contentBytes, _ := ioutil.ReadFile(*paramFile)
			contentBytes = append(contentBytes, []byte(tasksString)...)

			ioutil.WriteFile(*paramFile, contentBytes, fs.FileMode(os.O_APPEND))
			fmt.Println("Writing to", file.Filename)
		}
	}
}

////// debugging

func x(o interface{}) string {
	b, _ := json.MarshalIndent(o, "    ", "")
	return string(b)
}
func px(o interface{}) { fmt.Println(x(o)) }
