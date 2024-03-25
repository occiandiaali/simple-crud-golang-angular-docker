# Simple CRUD Project with Golang, Angular & Docker  

## Start  
Open your terminal and cd to the project, in the root folder run:  
`docker compose up`  

Open a new tab in terminal, cd to backend folder and run:  
`docker compose build`  
`docker compose up goapp`  

## Troubleshoot  
If error on first command run like:  
`Error response from daemon: driver failed programming external connectivity on endpoint db (9de2917ca5d13a9f5296fc5173b9897f9494706703c36b663fb7fce636f0fada): Error starting userland proxy: listen tcp4 0.0.0.0:5432: bind: address already in use`  
Run: `sudo lsof -i -P -n | grep LISTEN` to find the PID of the process using the port, then run `kill -9 <PID>` (you may need to add 'sudo' b4 kill)  
