## Goal 
> This repository contains a simple server implementation of the Pokdeng card game using Golang and TCP connections. The goal of this project is to demonstrate basic server-side game logic using plain text communication and incorporating concurrency concepts such as mutexes, channels, atomic values, peer connection handling and multiple reader and writer in IO concept.



For more information about how to play Pokdeng, visit this <a style="color:blue;" href="https://www.coololdgames.com/card-games/gambling/pok-deng/">link</a>.


### A Command that can execute to the server 
| Command             | Description                                                                                                                                                   |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| /help               | Get all the command that you can execute to server                                                                                                            |
| /create-room        | Create a room and you will be a dealer, you can create only one room per timeto create a room and you will be a dealer, you can create only one room per time |
| /list-room          | List all the rooms that you can join                                                                                                                          |
| /join-room (roomId) | Join the room with roomId                                                                                                                                     |
| /current-room       | Show the current toom that you are exist                                                                                                                      |
| /start              | Start the game, only the dealer (player1) have a privilege to start the game                                                                                  |
| /current-game       | Show the current status of the game, current hand and which hand are you                                                                                      |
| /stay               | If you are satisfy with two cards you can choose stay                                                                                                         |
| /more               | if you are not satisfy with two cards you ask for an extra card                                                                                               |
| /close-room         | Close the room and every player will get leave the room automatically, only the dealer (player1) have a privilege to close the room                           |


### Contributing
Me welcome contributions to the Pokdeng server project! Please fork the repository and submit a pull request with your changes. Make sure to follow the coding style and include tests for new features.

