# fetch-assignment

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Usage](#usage)
    - [Running the Server](#running-the-server)
    - [API Endpoints](#api-endpoints)
4. [Testing](#testing)
5. [Project Structure](#project-structure)
6. [Contributing](#contributing)
7. [License](#license)

## Introduction
A simple receipt processor that awards points based on various rules.

Done for Fetch take home assessment. To be closer to production ready, I would probably replace the built in MUX with something more robust, but for this project there was no specifications around request logging or authentication.

I tried to cover every use case from the source repo with unit tests. I am sure I missed a few. 

## Installation
1. Clone the repository:
    ```sh
    git clone https://github.com/keith-decker/fetch-assignment.git
    cd fetch-assignment
    ```
2. Install dependencies:
    ```sh
    go mod tidy
    ```
3. Run the project: 
    ```sh
    go run main.go
    ```

## Usage
See api.yml for a description of the endpoints

### Running the Server
use -port to modify the default port from 8080
```sh
go run main.go -port 9090
```

### API Endpoints
See api.yml

## Testing
```sh
go test ./...
```

## License
MIT. Though you probably don't really want to use this elsewhere. It needs some more work to be production ready.
