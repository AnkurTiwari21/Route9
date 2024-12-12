# Route9 - Custom DNS Server in GO

A custom DNS server written in Go, designed for handling DNS queries efficiently and providing enhanced flexibility for learning and experimentation.

<img width="884" alt="Screenshot 2024-12-12 at 12 09 16 AM" src="https://github.com/user-attachments/assets/4f3764e1-48f1-4515-8219-765bf198de24" />

# Features
 - Resolves DNS queries using custom logic and response headers.
 - It also supports DNS Caching (with default TTL of `60sec`)
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

# v1.1 - Added DNS Caching and TTL (Time to Live)
 - Added an in-memory database(redis) for caching the query result.
 - Please make a `.env` file in the root with "REDIS_ADDRESS" AND "REDIS_PASSWORD" (see the .env)
 - RESULTS AFTER CACHING:
   - <img width="655" alt="Screenshot 2024-12-12 at 12 02 11 PM" src="https://github.com/user-attachments/assets/9c6f32eb-79bc-46d8-9093-8dbceb26dea5" />
  
   - The above SS shows 2 dig request to our LOCAL DNS server. At the time of first request, there was no data of `google.com` in the Local Cache and Response Time was ~`38ms`.
   - The Second time we already have the information of `google.com` cached locally so no `LOOKUP` is performed and the response time is REDUCED to ~`8ms`

 - For now, the TTL( Time to Live) is hardcoded with value `60sec`.   


# Contributing
Contributions are welcome! Feel free to submit issues or pull requests to enhance the project.
