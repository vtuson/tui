# tui
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/vtuson/tui)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/vtuson/tui/blob/master/LICENSE)

This is a Basic Implementation of a text ui that can be used to run CLI commands or you can define your own HandlerCommand to run Go functions
  Basic Structure is:
	Menu
 		[]Command
 			[]Args
 By default a CLI command handler is provided that is able to pass parameters as flag, options or Enviroment variables
 You can find an example app in the /example folder
