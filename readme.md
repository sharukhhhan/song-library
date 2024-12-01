# Song Library Service

The **Song Library Service** is a Go-based application for managing songs, including their metadata and lyrics. It provides features for filtering, pagination, and integration with an external API to enrich song data.

---

## **Key Features**

### 1. **Song Management**
- Add, update, retrieve, and delete songs in the library.
- Search songs using filters such as:
    - Title
    - Group name
    - Release date range
    - Link
    - Text (contains keyword)

### 2. **Lyrics Management**
- Paginate through song lyrics verse by verse.

### 3. **External API Integration**
- Fetch additional song details (release date, lyrics, and link) from an external API when adding a new song.
- Ensure the external API URL is specified in `configs.yaml` under the `ExternalAPI.URL` field.

---

## **How to Run the Project**

### **1. Set Up the Database**
- Ensure PostgreSQL is running and matches the configuration in `configs.yaml`.
- Automigrations are included and will run on project startup, creating necessary tables.

### **2. Start the Server**
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd song-library-service
   ```
2. Install dependencies
    ```bash
   go mod tidy 
   ```
3. Run the project
    ```bash
   go run cmd/main.go
   ```

### **3. API Documentation**
Swagger is integrated and available at:
```bash
http://localhost:8080/swagger/index.html
   ```