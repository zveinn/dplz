# Dplz
Dplz is a deployment system that only requires BASH and SSH. No installation is required on the remote machine in order for it to work. <br> <br>
Dplz is capable of running shell commands and copying any type of file or directory contents. Dplz will also output a report that will tell the user if commands worked or failed.
 <br> <br>
 Right now this software is in beta, but I would love for anyone and everyone to try it out and flood me with issues to work on.

# On the horizon
0. Basic tests
1. More advanced filtering
2. Verbose output mode
3. Pre/Post script hooks
4. [insert your suggestion here]

# How to install
```bash
# Download a pre-built binary (linux/macos) from the releases page
# https://github.com/zveinn/dplz/releases

# .. or build from source
$ git clone https://github.com/zveinn/dplz
$ cd dplz
$ go build -o dplz .
```

# How to get started
1. [Define a server](#Defining-servers)
    - Server files can be located anywhere you please
2. [Define variables (optional)](#Defining-variables)
    - Variable files can be located anywhere you please
3. [Define a script](#Defining-scripts)
    - Paths inside scripts are resolved relative to the directory you run dplz from
4. [Run a deployment](#Running-deployments)

# Running deployments
```bash
# Just run a basic deployment
$ dplz -servers=[PATH_TO_YOUR_SERVER_FILE] -script=[PATH_TO_YOUR_SCRIPT_FILE]
# Run a deployment with variables
$ dplz -servers=server.json -script=script.json -vars=vars.json
# The same deployment using shorthand flags
$ dplz -s server.json -sc script.json -v vars.json
# Run a deployment and ignore the warning prompt
$ dplz -s server.json -sc script.json -ignorePrompt
# Run a deployment, ignore the warning prompt and use a filter.
$ dplz -s server.json -sc script.json -ignorePrompt -filter "script.cmd"
```
Before running, dplz prints a summary of the server, script and variable files and waits for confirmation. Press `N` (or `n`) to cancel, anything else continues. Use `-ignorePrompt` to skip this.

# Command Line Arguments
The `servers` and `script` flags are required, everything else is optional.
```go
	flag.String("script", "", "The path to your script file")               // shorthand: -sc
	flag.String("servers", "", "The path to your server file")              // shorthand: -s
	flag.String("vars", "", "The path to your variables file")              // shorthand: -v
	flag.String("filter", "", "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ") // shorthand: -f
	flag.Bool("ignorePrompt", false, "Add this flag to skip the confirmation prompt")
```

# Defining servers
Before you define any deployments you need to define a target server. <br>
Each server object contains some basic connection information and custom variables. <br>
- Variables are accessible in commands and templates as `{[server.variables.NAME]}`
- Authentication is done with a private key (`key`) or a password (`password`). If both are set, the key is used.
- Host key verification is skipped and the connection timeout is 10 seconds.
```json
{
    "hostname": "googlecloud-dev-01",
    "ip": "131.161.181.11",
    "port": "22",
    "key": "/home/user/.ssh/ssh-key",
    "password": "",
    "user": "root",
    "variables": {
        "privateIP": "11.11.11.11",
        "dns": "1.1.1.1"
    }
}
```

# Defining deployments
Deployment files have been removed. The server, script and variable file paths are now given directly as command line arguments, which makes a deployment fully described by the command that runs it:
```bash
$ dplz -s cloud-dev.server.json -sc basic.script.json -v dev.variables.json
```

# Defining variables
Variables are defined in a json file and the `vars` flag is used to load variables each time you run a deployment.
- These variables are available inside commands and templates.
```json
{
    "testVariable": "Dev variables loaded",
    "redisIP": "198.168.0.66",
    "sqlIP": "192.168.0.67"
}
```
Variables are injected into `run` commands, `file` paths, `template` paths and the contents of template files using the `{[name]}` syntax. The following variables are always available:
- `{[NAME]}` .. any variable from the variables file
- `{[server.ip]}` `{[server.port]}` `{[server.user]}` `{[server.key]}` `{[server.hostname]}`
- `{[server.variables.NAME]}` .. variables from the server file
- `{[script.variables.NAME]}` .. variables from the script file
- `{[script.tag]}` .. the script name
- `{[deployment.servers]}` `{[deployment.project]}` `{[deployment.varFile]}` .. the paths given on the command line (project = script)

# Defining scripts
This is a basic script, it will contain some variables and commands to execute. There are four types of commands:
1. Run - Runs a command on the server
2. File - Copies a file AS IS to the server (follows SCP syntax)
3. Template - Copies a file AND replaces variables (follows SCP syntax)
4. Directory - Recursively copies the contents of a directory to the server (follows SCP syntax)

Every command supports these options:
- `filter` .. a tag used to select the command with the `-filter` flag
- `async` .. run the command in the background, ordering is no longer guaranteed
- `local` .. run the command on the local machine (using `sh -c`) instead of the server
- `skip` .. skip the command by default; it only runs when a `-filter` matching it is given

NOTE: `file` and `template` use `local`/`remote` keys, while `directory` uses `src`/`dst`.

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
                "local": "files/webserver.template.conf",
                "remote": "/home/sveinn/meow/test.template",
                "mode": "0777"
            },
            "filter": "templates",
            "async": true
        },
        {
            "file": {
                "local": "files/meow",
                "remote": "/home/sveinn/meow/test.file",
                "mode": "0777"
            },
            "filter": "files",
            "async": true
        },
        {
            "run": "mkdir /home/sveinn/meow",
            "filter": "dir",
            "async": false
        },
        {
            "directory": {
                "src": "files",
                "dst": "/home/sveinn/meow",
                "mode": "0777"
            },
            "filter": "dir",
            "async": false
        },
        {
            "run": "uname -a",
            "filter": "uname",
            "async": false
        },
        {
            "run": "echo 'this only runs locally'",
            "filter": "local",
            "local": true
        },
        {
            "run": "echo 'this only runs when a filter selects it'",
            "filter": "manual",
            "skip": true
        },
        {
            "run": "ls -la",
            "filter": "list"
        }
    ]
}
```

# Loops
Commands and file copies can be expanded into loops using the `{[start..end]}` syntax. The range is inclusive.

Loop over a command (expands into a single `cmd1 && cmd2 && ..` execution):
```json
{ "run": "touch /home/sveinn/file-{[1..5]}" }
```

Copy one local file to many remote names:
```json
{ "file": { "local": "files/meow", "remote": "/home/sveinn/meow-{[1..3]}", "mode": "0777" } }
```

Copy many local files to many remote names (both ranges are paired index by index):
```json
{ "file": { "local": "files/meow-{[1..3]}", "remote": "/home/sveinn/meow-{[1..3]}", "mode": "0777" } }
```
Loops work in `run` commands (remote and local) and in `file` copies. A loop on only the local side of a `file` copy is not supported.

# Ordering of scripts
A single script file is loaded per run. The commands within it are executed against every server, one server at a time.

# Ordering of commands
Commands present inside scripts will always be executed in order. Except if the async tag is specified, then ordering is not guaranteed for the commands flagged as async.

# Filtering
The `-filter` flag takes two parts separated by a dot: `SCRIPT_FILTER.CMD_FILTER`. The first part is matched against the `filter` field of the script, the second against the `filter` field of each command. Matching is strict, but either part can be replaced with `*` to match everything.

 Run a certain script and commands that match the filter
 -  `-filter script.cmd`

 Run a certain set of commands inside all scripts
 - `-filter *.cmd`

Run all commands inside a script matching the filter
 - `-filter script.*`

Commands marked with `"skip": true` are only executed when a filter selecting them is given.
