# smock [![Build Status](https://travis-ci.org/vvatanabe/smock.svg?branch=master)](https://travis-ci.org/vvatanabe/smock)
simple mock generator

## Description
Automatically generate simple mock code of Golang interface.

## Installation
This package can be installed with the go get command:
```
$ go get github.com/vvatanabe/smock/cmd/smock
```

Built binaries are available on Github releases: https://github.com/vvatanabe/smock/releases

## Usage
```
$ cat $INPUT_FILE_PATH | smock -pkg=mock > $OUTPUT_FILE_PATH 
```

## Bugs and Feedback
For bugs, questions and discussions please use the Github Issues.

## License
[MIT License](http://www.opensource.org/licenses/mit-license.php)

## Author
[vvatanabe](https://github.com/vvatanabe)