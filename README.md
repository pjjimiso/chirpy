#### psql connection string: 
psql "postgres://postgres:postgres@localhost:5432/chirpy"

#### Create User
curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email": "test@email.com"}'

#### User Login
curl -X POST http://localhost:8080/api/login -H "Content-Type: application/json" -d '{"password": "password123", "email": "pjjimiso@email.com", "expires_in_seconds": "120"}'

#### Refresh Access Token (returns JWT)
curl -X POST http://localhost:8080/api/refresh -H "Authorization: Bearer d06cff4e0c2c0fae70da7dc899a86a53bb46c6187efab5733cc97e2b481dee20"

#### Create Chirp
curl -X POST http://localhost:8080/api/chirps -H "Content-Type: application/json" -d '{"body": "Hello, world!", "user_id": "e00b789e-67ac-4533-a98d-658aa583238f"}'
##### with jwt:
curl -X POST http://localhost:8080/api/chirps -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiJmN2I4ZDMwNS0zO" -d '{"body": "Hello, world!", "user_id": "e00b789e-67ac-4533-a98d-658aa583238f"}'

#### Reset API metrics and truncate users table
curl -X POST http://localhost:8080/admin/reset

