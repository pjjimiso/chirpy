#### psql connection string: 
psql "postgres://postgres:postgres@localhost:5432/chirpy"

#### Send POST request with JSON data
curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email": "pat.jimison@gmail.com"}'

#### Reset API metrics and truncate users table
curl -X POST http://localhost:8080/admin/reset
