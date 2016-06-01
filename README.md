# CatchARide-API

## What is CatchARide?

CatchARide was a Cordova/Angular app built as a proof of concept for a University of Washington Info 200 class. Its purpose was helping University of Washington students who commute to campus find groups to carpool with.

A working demo is hosted at http://www.catcharide.today

This repository holds the back-end code for the app. The angular front-end can be found [here](https://github.com/DaltonCaughell/CatchARide-App)

## What is CatchARide’s back-end stack?

* MySQL
* Docker
* Golang
* GORM
* Martini

## Can I build/run CatchARide?

CatchARide was created as a proof of concept for a demo day. There may be bugs and or security flaws. I would not recommend using CatchARide as is. However, CatchARide can be built by running `go get . && go run server.go` which will start it listening for connections on port 3000 by default.
