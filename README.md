# Route9 - Custom DNS Server in GO

A custom DNS server written in Go, designed for handling DNS queries efficiently and providing enhanced flexibility for learning and experimentation.

<img width="884" alt="Screenshot 2024-12-12 at 12 09 16 AM" src="https://github.com/user-attachments/assets/4f3764e1-48f1-4515-8219-765bf198de24" />

# Features
 - Resolves DNS queries using custom logic and response headers.
 - Implements both request parsing and response generation.
 - Supports UDP protocol for fast query handling.
 - Built-in error handling and logging for debugging.

# Prerequisites
 - Go 1.18 or later

# Installation
 1. Clone the repository
 2. build the main.go (go build app/main.go)
 3. RUN `./main --resolver <PASS ANY DNS RESOLVER IN <IP>:<PORT> FORMAT`
    - example : ./main.go --resolver 8.8.8.8:53

# Usage
 - By default, the server listens on 127.0.0.1:2053.
 - Use tools like dig to test queries:
   - `dig @127.0.0.1 -p 2053 google.com`
<img width="504" alt="Screenshot 2024-12-12 at 12 09 37 AM" src="https://github.com/user-attachments/assets/f443aaa4-d673-4464-a38a-0a0f39b26c8c" />



# Contributing
Contributions are welcome! Feel free to submit issues or pull requests to enhance the project.
