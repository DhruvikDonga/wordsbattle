
# WordsBattle

MiniWordGames is an engaging multiplayer word game that combines the thrill of real-time battles with the challenge of word creation. This project utilizes Golang for the backend server, Vue.js for the frontend application, WebSocket for real-time communication, and Docker for easy deployment. The game is hosted under the domain [miniwordgames.com](https://miniwordgames.com).

## Tech Stack  
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white) ![Vue.js](https://img.shields.io/badge/vuejs-%2335495e.svg?style=for-the-badge&logo=vuedotjs&logoColor=%234FC08D) ![Vuetify](https://img.shields.io/badge/Vuetify-1867C0?style=for-the-badge&logo=vuetify&logoColor=AEDDFF) ![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white) 	![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white) 

## Features

- **Real-time Battles:** Challenge your friends or random opponents to fast-paced word battles.
- **WebSocket Integration:** Enjoy seamless and instant communication between players for a responsive gaming experience.
- **Golang Backend:** Utilizes the power of Golang to handle server-side logic efficiently.
- **Vue.js Frontend:** A dynamic and interactive user interface designed with Vue.js for a smooth gaming experience.
- **Dockerized Multistaged Deployment:** Easily deploy and manage the application using Docker containers which are small and lightweight.


## Architecture  
```mermaid
graph TD
    A[Vue Client] --> B[Go Server]
    B --> C[HTTP server]
    B ==> D[Websocket client upgrader]
    D ==> E[Concurrent client reader]
    D --> F[Concurrent client writer]
    B ==> G[Game Server]
    G -.->|manage the rooms and clients <br> Manage the lobby for random rooms| H[Room Server <small><i>id:-axGrw</i></small>]
    E -->|Usecases <br>Lobby server for random games connect two users create a room|G
    G -->|Usecases <br>Give user notificatons incase of room is not connected due to reasons|F
    B --> H
    E ==>|send the data from client to room server| H
    H ==> F
    F ==>|Send data to all the clients in list| A
    H <--> I[Game State Manager <br> <small><ul><li>Game algorithm</li><li>Send Game State to the <br>UI client on time ticker</li><li><b>Process the data of the user</b></li></ul> </small>]
    H -.->|Send game state data <br> <-ticker.C to all clients in a room|F
```
## Demo

https://youtu.be/-9HrFUU_jfs?si=RhmSZE6-uSq-6M8f  
Live :- http://miniwordgames.com

## Project Setup
wordsbattle uses Docker for deployment and project creation you can find docker-compose.yaml  in repo .

```
git clone git@github.com:DhruvikDonga/wordsbattle.git
cd miniwordgames

docker-compose build
docker-compose up
```

## Interested to contribute 
Checkout this issue :- https://github.com/DhruvikDonga/wordsbattle/issues/12 
Dev Branch :- https://github.com/DhruvikDonga/wordsbattle/tree/WB-8

## Authors

- [@Dhruvik D.](https://www.github.com/DhruvikDonga)

