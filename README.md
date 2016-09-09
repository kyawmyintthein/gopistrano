=======
gopistrano
==========

Automatic Deployment Tool for Beego + Tmux based on github.com/alanchavez88/gopistrano

## Requirements
* GoLang >= 1.3.3

## Installation
Run this if you want to install gopistrano binary

``` go
go get github.com/sangchul-sim/gopistrano
```

That will compile and install gopistrano in your deploy config

To deploy a project, you need to create a deploy config. A deploy config is just a configuration file in TOML that will contain the credentials to SSH into your server, and the path of the directory where you want to deploy your project to.

This is a sample TOML config file
```
[login]
username = "yourusername"
password = "yourpassword"
ssh_path = ""

[servers]
    [servers.development]
        ip = ["10.0.0.1", "10.0.0.2"]
        port = 22

    [servers.staging]
        ip = ["10.0.0.3", "10.0.0.4"]
        port = 22

    [servers.production]
        ip = ["10.0.10.1", "10.0.10.1"]
        port = 22

[deploy]
repository = "https://github.com/sangchul-sim/webapp-golang-beego.git"
go_project_path = "/home/yourusername/go_project"
package_name = "github.com/sangchul-sim/webapp-golang-beego"
app_name = "webapp-golang-beego"
keep_releases = 5
use_sudo = false
webserver_user = "yourusername"
```
The file above will clone the git repository above into the path specified in the TOML config file. 

It use https://tmux.github.io/ to run beego project in background. Tmux use current path to run the beego project in background.
But, need to do cross-compiles the beego project before deploy. Next version will support auto compile.


To deploy you have to run

``` sh
gopistrano -config=conf/deploy.conf -env=development -action=setup

```

and then:

``` sh
gopistrano -config=conf/deploy.conf -env=development -action=deploy
```

## Support

Need help with Gopistrano? email at treedy@gmail.com
