## Draw2Gather

This repository contains API's and game handling logic of Draw2Gather. It is written in Go and uses Firebase Firestore database and Firebase Authentication, therefore it requires to connect to a Firebase project. Before setup you should install [Go](https://go.dev/dl/), [NodeJS](https://nodejs.org/en/download/current), and [Firebase CLI](https://www.npmjs.com/package/firebase-tools). Firebase CLI requires Java 11 or later.

### Configuration

This application requires a Firebase project. After a Firebase project is created 2 services also needs to be configured. It is easy to do in Firebase console, but emulators can also be used. It's your decision. Setup below uses Firebase emulators instead of configuring services.

Also API's CORS setting might need to be configured. In demo branch it's configured to allow origin ```http://localhost:5173``` this origin is default when you run our [website](https://github.com/INF303Project/draw2gather-web), if you run it on another port also change the CORS configuration at https://github.com/INF303Project/draw2gather/blob/demo/internal/api/handler.go#L72. Also this configuration means that only the clients in your machine can access to API. If you want to make it available to all devices in your local network change localhost with your IP address or add it as a second origin.

Note: API uses session cookies for authentication and our game does not allow a user (browser client really) to play more than one game at a time. Thus when you want to test the game, you should use two different browsers or a private window for second user.

### Setup

Main branch is for production, right now it's available at [api.draw2gather.online](https://api.draw2gather.online/health).

- Clone the repository
    ```
    git clone https://github.com/INF303Project/draw2gather.git
    ```

- Checkout to demo brach
    ```
    cd draw2gather
    git checkout demo
    ```

- Login to Firebase (on Windows you might need to change ExecutionPolicy)
    ```
    firebase login
    ```

- Create a new Firebase project
    ```
    firebase projects:create <<name>>
    firebase use <<name>>
    ```

- Goto Firebase console, in project page, goto project settings. Under service accounts create a new private key. Rename it to ```admin-sdk.json``` and move it under this directory.

- Create a ```.env``` file with content
    ```
    FIREBASE_AUTH_EMULATOR_HOST="127.0.0.1:9099"
    FIRESTORE_EMULATOR_HOST="127.0.0.1:9100"
    ```

- Start Firebase emulators (this will install emulators if necessary)
    ```
    firebase emulators:start
    ```

- Initialize default word sets
    ```
    go run ./cmd/words/main.go
    ```

- Start API
    ```
    go run ./cmd/draw2gather/main.go
    ```

- Start website (see [draw2gather-web](https://github.com/INF303Project/draw2gather-web))
