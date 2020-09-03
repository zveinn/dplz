# Dplz
Dplz is a deployment system that only requires BASH and SSH. No installation is required on the remote machine in order for it to work. <br> <br>
Dplz is capable of running shell commands and copying any type of file or directory contents. Dplz will also output a report that will tell the user if commands worked or failed. 
 <br> <br>
 Right now this software is in beta, but I would love for anyone and everyone to try it out and flood me with issues to work on.

# On the horizon
1. Localhost support
2. More advanced filtering
3. Verbose output mode
4. [insert your suggestion here]

# How to install
```bash
# Get the project using go get
$ go get github.com/zveinn/dplz
# use go to install
$ go install .
```

# How to get started
1. [Define a server](#Defining-servers)
2. [Define a deployment (optional)](#Defining-deployments)
3. [Define variables](#Defining-variables)
4. [Define a script](#Defining-scripts)
5. [Run a deployment](#Defining-a-server)

# Running deployments 
```bash
# Just run a basic deployment
$ dplz -deployment=[PATH_TO_YOUR_DEPLOYMENT_FILE] 
# Run a deployment and ignore warning prompt
$ dplz -deployment=[PATH_TO_YOUR_DEPLOYMENT_FILE] -ignorePrompt 
# Run a deployment, ignore the warning prompt and use a filter.
$ dplz -deployment=[PATH_TO_YOUR_DEPLOYMENT_FILE] -ignorePrompt -filter "script.cmd"

# Run a deployment using command line arguments only.
$ dplz -servers=[PATH_TO_SEVER_FOLDER] -project=[PATH_TO_PROJECT_FOLDER] -vars=[PATH_TO_VARIABLES_FILE]
```


# Command Line Arguments
Command line arguments can be combines with deployment json files. If you do combine them, the command line arguments will overwrite the deployment file settings.
```go
	flag.String("deployment", "", "The path to your deployment file")
	flag.String("project", "", "The path to your project files (not needed if using a deployment file)")
	flag.String("servers", "", "The path to your server files (not needed if using a deployment file)")
	flag.String("vars", "", "The path to your variables file (not needed if using a deployment file)")
	flag.Bool("ignorePrompt", false, "Add this flag to skipt the confirmation prompt")
	flag.String("filter", "", "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ")
```

# Defining servers
Before you defined any deployments you need to define a target server. <br>
Each server object contains some basic connections information, custom variables and PRE/POST scripts. <br>
- Variables are accessable in commands and tamplates. 
- Pre scripts will run BEFORE any other scripts run on that server
- Post scripts will run AFTER all scripts have been executed.
```json
{
    "hostname": "googlecloud-dev-01",
    "ip": "131.161.181.11", 
    "port": "22", 
    "key": "/home/user/.ssh/ssh-key", 
    "user": "root", 
    "variables": { 
        "privateIP": "11.11.11.11",
        "dns": "1.1.1.1"
    },
    "pre": [ 
        {"run": "echo 'This runs before all scripts'"}
    ],
    "post": [ 
        {"run": "echo 'This runs after all script'"}
    ]
}
```

# Defining deployments
This part is purely optional, you can input all of the below parameters are command line arguments. But for the sake of replicating deployments we decided to add this "highest level object". <br>
A Deployment will define the location of your server json files, project json files and the variables json file. 
- NOTE: Deployment wide variables are only configurable inside a deployment json file
```json
{
    "servers": "PATH_TO_YOUR_SERVER_FOLDER",
    "project": "PATH_TO_YOUR_PROJECT_FOLDER",
    "vars": "PATH_TO_YOUR_VARIABLES_FILE",
    "variables": {
        "example": "This is an example custom variable"
    }
}
```

# Defining variables
Variables are defined in a json.file and the -vars flag is used to load variables each time you run a deployment.
- These variables are available inside commands and templates. 
```json
{
    "testVariable": "Dev variables loaded",
    "redisIP": "198.168.0.66",
    "sqlIP": "192.168.0.67"
}
```

# Defining scripts
This is a basic script, it will contain some variables and some commands to execute. There are three types of commands:
1. Run - Only runs a commands on the server
2. File - Copies a file AS IS to the server ( follows SCP syntax )
3. Template - Copies a file AND replaces variables ( follows SCP syntax )

```json
{
    "name": "Basic deployment",
    "filter": "basic",
    "variables": {
        "name": "Some random name",
        "testVariable": "Here is some random text you can inject into a file..."
    },
    "cmd": [
        {
            "template": {
                "local": "projects/dplz/files/webserver.template.conf",
                "remote": "test.template"
            },
            "filter": "templates",
            "async": true
        },
        {
            "file": {
                "local": "projects/dplz/files/webserver.template.conf",
                "remote": "test.file"
            },
            "filter": "files",
            "async": true
        },
        {
            "run": "uname -a",
            "filter": "uname",
            "async": true
        },
        {
            "run": "ls -la",
            "filter": "list"
        }
    ]
}
```

# Ordering of scripts
Scripts do not have any guarentee to be ran in a particular order. The ordering is soly based on the directory walking machanism, which makes it non-dependable.

# Ordering of commands
Commands present inside scripts will always be executed in order. Except if the async tag is specified, then ordering is not guaranteed for the commands flagged as async.

# filtering
The filtering is currently a strict matching filter. The filter tag and the filter variables on the script or command need to match exactly.

 Run a certain script and commands that match the filter
 -  -filter scripts.cmd

 Run a certain set of commands inside all filters
 - -filter *.cmd

Run all commands inside a scripts matching the filter
 - -filter scripts.*

