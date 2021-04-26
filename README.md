## Simple Task manager for markdown files

### Setup

Checkout the repository beside your dendron workspace folder

```
|-workspace/
|   |-vault/
|       |-.vscode/
|           |-tasks.json
|           
|-dendronutils/
    |-cmd/
    |   |-gettasks
    |-build.sh

```

### Build

All dependencies are vendored to make builds easy-peasy.

Simply run 

```
./build.sh
```

### Usage

```
Usage of ./gettasks:
  -file string
    	current file
  -hirearchy string
    	filter tasks from a hirearchy (default "daily.journal")
  -write
    	write to selected file
```

## Configure VSCode tasks

Add the following to `.vscode/tasks.json`
Note: Replace `../gettasks` with the relative or absolute path to the binary

```
{
    // See https://go.microsoft.com/fwlink/?LinkId=733558
    // for the documentation about the tasks.json format
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Get open tasks",
            "type": "shell",
            "command": "../../dendronutils/gettasks -file ${file} -write true",
        }
    ]
}
```

### Run VSCode Task

1. `Cmd+Shift+P` 
2. `Run Task` 
3. `Get Open tasks`

