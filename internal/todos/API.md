# Todo API Documentation

This document describes the RESTful API endpoints available for managing todo items.

## Data Model

### TodoItem

```json
{
  "id": "uint",
  "text": "string",
  "done": "boolean"
}
```

## Endpoints

### Get All Todos

Retrieves a list of all todo items.

- **URL**: `/todos`
- **Method**: `GET`
- **Auth Required**: No
- **Success Response**:
  - **Code**: 200 OK
  - **Content**: Array of TodoItem objects
- **Error Response**:
  - **Code**: 500 Internal Server Error
  - **Content**: `{ "error": "could not read todo-items" }`

### Create Todo

Creates a new todo item.

- **URL**: `/todos`
- **Method**: `POST`
- **Auth Required**: No
- **Request Body**:
  ```json
  {
    "text": "string",
    "done": false
  }
  ```
- **Success Response**:
  - **Code**: 200 OK
  - **Content**: Created TodoItem object
- **Error Response**:
  - **Code**: 400 Bad Request
  - **Content**: `{ "error": "could not create todo-item" }`

### Update Todo

Updates an existing todo item by ID.

- **URL**: `/todos/:id`
- **Method**: `PUT`
- **Auth Required**: No
- **URL Parameters**:
  - `id`: The ID of the todo item to update
- **Request Body**:
  ```json
  {
    "text": "string | null",
    "done": "boolean | null"
  }
  ```
  Note: Both fields are optional. Only provided fields will be updated.
- **Success Response**:
  - **Code**: 200 OK
  - **Content**: Updated TodoItem object
- **Error Response**:
  - **Code**: 400 Bad Request
  - **Content**: 
    - `{ "error": "invalid id" }` - If ID is invalid
    - `{ "error": "could not update todo-item" }` - If update fails

### Delete Todo

Deletes a todo item by ID.

- **URL**: `/todos/:id`
- **Method**: `DELETE`
- **Auth Required**: No
- **URL Parameters**:
  - `id`: The ID of the todo item to delete
- **Success Response**:
  - **Code**: 200 OK
  - **Content**: Empty string
- **Error Response**:
  - **Code**: 400 Bad Request
  - **Content**: 
    - `{ "error": "invalid id" }` - If ID is invalid
    - `{ "error": "could not delete todo-item" }` - If deletion fails

## Error Handling

All endpoints return appropriate HTTP status codes and error messages in the following format:

```json
{
  "error": "error message description"
}
```