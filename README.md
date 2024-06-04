1. Peer connected to server
1.1 Check is peer in the list or not
1.2 Send encrypt handshake 
1.3 Add peer using channel
1.4 In the read loop accept the channel of add peer
1.5 Handle adding new peers to list if any error occur close the connection of that peer instantly
2. Server listen to message from peer
2.1 Command that user can use
2.1.1 /create-room
2.1.2 /list-room
2.1.3 /join-room/(roomId)
2.1.4 /delete-room/(roomId) only owner can delete and also delete every room if the owner of the room disconnected
4. First peer will have starter money of 1000
5. Game state using atomic value to change the game state


/create-room
/join-room 123