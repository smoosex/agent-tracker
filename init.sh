#!/bin/bash

# Setup backend
cd server
go mod tidy
cd ..

# Setup frontend
cd web
npm install
cd ..

# Initialize data directory
mkdir -p data

echo "Setup complete!"
echo "To start the backend: cd server && go run main.go"
echo "To start the frontend: cd web && npm run dev"