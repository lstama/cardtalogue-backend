# Card Recognition and Image Upload API Documentation

## Overview

The image upload functionality now includes optional card recognition through a third-party API. When enabled, it can identify cards from images and return detailed match information.

## Configuration

Add the following environment variable to enable card recognition:

```env
CARD_SCAN_API_URL=http://localhost:5000/api/recognize
```

If `CARD_SCAN_API_URL` is not set or empty, the recognition feature will be disabled and the system will work normally without scanning.

## API Endpoints

### Upload Image with Optional Recognition

**Endpoint:** `POST /api/v1/media/upload`

**Authentication:** Required (Bearer token, editor or admin role)

**Parameters:**
- `file` (form-data, required): Image file (JPEG, PNG, GIF)
- `also_scan` (query, optional): Set to `true` to enable card recognition. Defaults to `false`

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer <token>" \
  -H "Accept: application/json" \
  -F "file=@card.jpg" \
  -F "also_scan=true"
```

**Response - Without Scan (also_scan=false or not set):**
```json
{
  "data": {
    "id": 1,
    "url": "http://localhost:8080/media/abc123_def456.jpeg",
    "status": "attached",
    "provider": "local",
    "created_at": "2024-01-20T12:00:00Z",
    "updated_at": "2024-01-20T12:00:00Z"
  }
}
```

**Response - With Scan (also_scan=true):**
```json
{
  "data": {
    "id": 1,
    "url": "http://localhost:8080/media/abc123_def456.jpeg",
    "status": "attached",
    "provider": "local",
    "created_at": "2024-01-20T12:00:00Z",
    "updated_at": "2024-01-20T12:00:00Z"
  },
  "scan": {
    "matches": [
      {
        "name": "Enlilgirsu, the Orcust Mekk-Knight",
        "confidence": 0.0372421875,
        "description": {
          "effect": null,
          "monster_effect": "2+ monsters, including an \"Orcust\" Link Monster...",
          "pendulum_effect": null
        }
      },
      {
        "name": "Blackwing - Sudri the Phantom Glimmer",
        "confidence": 0.0262109375,
        "description": {
          "effect": null,
          "monster_effect": "When this card is Normal Summoned...",
          "pendulum_effect": null
        }
      }
    ]
  }
}
```

### Serve Image

**Endpoint:** `GET /media/{filename}`

**Authentication:** Not required (public endpoint)

**Parameters:**
- `filename` (path, required): The filename returned from upload

**Response:** Binary JPEG image

**Example:**
```bash
curl http://localhost:8080/media/abc123_def456.jpeg -o downloaded.jpg
```

## Card Recognition Details

### How It Works

1. When `also_scan=true` is passed, the backend sends the uploaded image to the configured `CARD_SCAN_API_URL`
2. The third-party API processes the image and returns potential card matches
3. Matches are extracted with the following information:
   - **name**: The card name
   - **confidence**: A confidence score (0-1) indicating match certainty
   - **description**: Contains the card effects (effect, monster_effect, pendulum_effect)

### Important Notes

- **Non-blocking**: If card recognition fails, the image upload still succeeds
- **Graceful degradation**: If the recognition API is down or returns an error, a warning is logged but the operation continues
- **Optional feature**: The recognition feature is only enabled if `CARD_SCAN_API_URL` is configured
- **Processed data**: Only name, confidence, and description fields are extracted and returned to the client

## Error Handling

### Upload Errors

- `400 Bad Request`: File parameter missing
- `400 Bad Request`: Invalid filename (path traversal attempt)
- `500 Internal Server Error`: File processing failed, database error, or file system error

### Recognition Errors

Recognition errors do NOT cause the upload to fail. If the recognition API:
- Is not configured: No recognition performed, no error returned
- Is unreachable: Warning logged, upload succeeds without scan data
- Returns an error: Warning logged, upload succeeds without scan data
- Returns success=false: Warning logged, upload succeeds without scan data

## Image Processing

- **Formats supported**: JPEG, PNG, GIF
- **Output format**: All images converted to JPEG for consistency
- **Quality**: 85% JPEG quality (good balance between quality and file size)
- **Filename**: `{hash}_{uuid}.jpeg`
  - Hash: First 16 characters of SHA256 hash of file content
  - UUID: First 8 characters of generated UUID
  - Minimizes collisions while keeping filenames manageable

## File Storage

- **Default path**: `./uploads/media/`
- **Configurable**: Set `MEDIA_UPLOAD_PATH` environment variable
- **Public access**: Files are served via `/media/{filename}` endpoint
- **Directory creation**: Automatically created if it doesn't exist

## Example: Full Workflow

```bash
# 1. Upload image with card recognition
RESPONSE=$(curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" \
  -F "file=@mysterycard.png" \
  -F "also_scan=true")

# 2. Parse response
MEDIA_URL=$(echo $RESPONSE | jq -r '.data.url')
CARD_NAME=$(echo $RESPONSE | jq -r '.scan.matches[0].name')
CONFIDENCE=$(echo $RESPONSE | jq -r '.scan.matches[0].confidence')

# 3. Use the media URL in your card creation
curl -X POST http://localhost:8080/api/v1/cards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"CODE\",
    \"name\": \"$CARD_NAME\",
    \"media_id\": null,
    \"image_url\": \"$MEDIA_URL\",
    \"game_id\": 1,
    \"for_sale\": \"yes\"
  }"
```

## Technical Architecture

### Service Separation

- **MediaService**: Handles image upload, storage, and database operations
- **CardRecognitionService**: Encapsulates third-party API communication
- **MediaHandler**: HTTP request handling and routing

### Benefits

- **Separation of concerns**: Each service has a single responsibility
- **Testability**: Services can be tested independently
- **Flexibility**: Recognition can be enabled/disabled via configuration
- **Error isolation**: Recognition failures don't affect core functionality
