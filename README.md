#### psql connection string: 
psql "postgres://postgres:postgres@localhost:5432/chirpy"

#### Create User
curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email": "pjjimiso@email.com", "password": "password123"}'

#### User Login
curl -X POST http://localhost:8080/api/login -H "Content-Type: application/json" -d '{"password": "password123", "email": "pjjimiso@email.com", "expires_in_seconds": "3600"}'

#### Refresh Access Token (returns JWT)
curl -X POST http://localhost:8080/api/refresh -H "Authorization: Bearer <refresh_token>"

#### Create Chirp using JWT
curl -X POST http://localhost:8080/api/chirps -H "Content-Type: application/json" -H "Authorization: Bearer <access_token>" -d '{"body": "Hello, world!"}'

#### Delete Chirp
curl -X DELETE http://localhost:8080/api/chirps/<chirp_id> -H "Authorization: Bearer <access_token>"
 
#### Reset API metrics and truncate users table
curl -X POST http://localhost:8080/admin/reset

#### Update user credentials using access token
curl -X PUT http://localhost:8080/api/users -H "Content-Type: application/json" -H "Authorization: Bearer <access_token>"  -d '{"password": "123password", "email": "pjjimiso@gmail.com"}'

#### Get All Chirps
curl -X GET https://localhost:8080/api/chirps

#### Get Chirps by Author
curl -X GET http://localhost:8080/api/chirps?author_id=<user_id>

#### Get Chirp by id
curl -X GET http://localhost:8080/api/chirp/<chirp_id>
