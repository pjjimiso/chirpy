#### psql connection string: 
psql "postgres://postgres:postgres@localhost:5432/chirpy"

#### Create User
curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email": "test@email.com"}'

#### Create Chirp
curl -X POST http://localhost:8080/api/chirps -H "Content-Type: application/json" -d '{"body": "Hello, world!", "user_id": "e00b789e-67ac-4533-a98d-658aa583238f"}'

#### Reset API metrics and truncate users table
curl -X POST http://localhost:8080/admin/reset

