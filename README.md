# Online Gomoku Game

This project implements an online multiplayer Gomoku game. Gomoku is a board game where the first player to place five stones in a row horizontally, vertically, or diagonally wins. Two players take turns placing stones.

## Features

-   **Real-time Multiplayer Game**: Two players play the game in real-time over the internet.
-   **WebSocket Communication**: Real-time communication between the server and clients is achieved using WebSocket.
-   **Game Room Matching**: The server automatically matches two players to create a game room.

## Project Structure

This project consists of two parts:

1. **Server-side Code (Go)**: The code that implements the server for the online Gomoku game. It handles WebSocket connection management, game room matching, game logic, etc.

2. **Client-side Code (HTML, JavaScript)**: The client-side code that runs in web browsers. It constructs the user interface and communicates with the server in real-time using WebSocket to play the game.

## How to Run

1. Open Browser: To play the online Gomoku game, navigate to [stonify5.com](https://stonify5.com) in a web browser.

2. Play: Upon accessing through a web browser, two players will be matched, and the game will start. Each player takes turns placing stones aiming for victory.
