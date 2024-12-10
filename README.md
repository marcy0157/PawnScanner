
# PawnScanner

## Description
This project provides a system for managing and consulting data related to data breaches. It consists of two main programs:

1. **PwnScanner**: 
   - A frontend interface that allows users to check if an email has been involved in a data breach.
   - Connects to a MongoDB database to query data.
   - Displays where the breach occurred (e.g., Facebook, Twitter, etc.).

2. **PwnAdmin**:
   - An administration tool that allows uploading data breach files into the MongoDB database.
   - Useful for maintaining and updating data.

---

## Project Structure

![](?raw=true)
---

## How to Use the Project

### Prerequisites
- Docker and Docker Compose installed.
- (Optional) A configured MongoDB instance if using `composeNOMongo.yml`.

### Configuration
1. Download the chosen compose file.

2. Configure the files:
   - `composeMongo.yml`: Ensure the configurations match your local environment.
     ```
     services:
  pwnscanneradmin:
    image: stepsjr/pwnscanneradmin:1.0
    container_name: pwnscanneradmin
    ports:
      - "8081:8081"
    environment:
      ADMIN_USERNAME: [username]
      ADMIN_PASSWORD: [password]
      MONGODB_URI: mongodb://[usernameDB]:[passwdDB]@[IP]:[PORT]
      MONGODB_DBNAME: [database_name]
    networks:
      - pwnscanner-network

  pwnscanner:
    image: marci01/pwnscanner:1.0
    container_name: pwnscanner
    ports:
      - "8080:8080"
    environment:
      DB_TYPE: mongodb
      DB_HOST: mongo
      DB_PORT: 27017
      DB_USERNAME: [db_username]
      DB_PASSWORD: [db_password]
      DB_NAME: [database_name]
    networks:
      - pwnscanner-network

  mongo:
    image: mongo:5.0
    container_name: mongo
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: [root_username]
      MONGO_INITDB_ROOT_PASSWORD: [root_password]
    networks:
      - pwnscanner-network
    volumes:
      - mongo-data:/data/db

networks:
  pwnscanner-network:
    driver: bridge

volumes:
  mongo-data:
     ```
   - `composeNOMongo.yml`: Edit the connection data to the external MongoDB database.
     ```

     
     ```


---

### Starting the Containers

#### With MongoDB included
Use the `composeMongo.yml` file:
   docker-compose -f composeMongo.yml up

#### With external MongoDB
Use the `composeNOMongo.yml` file:
   docker-compose -f composeNOMongo.yml up

#### Stopping the Containers
To stop the containers, use:
   docker-compose down

---

## Main Features

### PwnScanner (Frontend)
- Checks if an email has been involved in a data breach.
- Displays details of each breach (e.g., the service involved).

### PwnAdmin (Admin Tool)
- Uploads breach files into the MongoDB database.
- Features to manage uploaded data.

---

## Additional Notes
- **Port**: Verify the ports exposed in the Docker Compose files and ensure they are not already in use.
- **Database**: If using `composeNOMongo.yml`, ensure the MongoDB database is correctly configured and accessible.
## Collaborators:
- https://github.com/StepsJr4
- https://github.com/Mirko1021
- https://github.com/Joghurtzz
- https://github.com/EgIx004
