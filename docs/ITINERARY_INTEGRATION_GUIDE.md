# Itinerary API - Frontend Integration Guide

## 📋 Table of Contents
1. [Overview](#overview)
2. [Authentication Requirements](#authentication-requirements)
3. [API Endpoints](#api-endpoints)
4. [Data Models](#data-models)
5. [Integration Workflow](#integration-workflow)
6. [User Journey & UI Flows](#user-journey--ui-flows)
7. [Error Handling](#error-handling)
8. [Performance Considerations](#performance-considerations)
9. [Best Practices](#best-practices)
10. [Testing Recommendations](#testing-recommendations)

---

## Overview

The Itinerary API allows users to create optimized travel routes through multiple places in Nashik. The system automatically calculates the best route based on distance, travel time, and selected transport mode.

### Key Features
- **Automatic Route Optimization**: Uses nearest-neighbor algorithm to find optimal visiting order
- **Multiple Transport Modes**: Walking, Driving, Cycling, Public Transit
- **Distance & Time Calculation**: Automatic calculation using Haversine formula
- **Flexible Visit Duration**: Customize how long to spend at each place
- **Real-time Updates**: Update itinerary details without recreating
- **Cascade Delete**: Deleting an itinerary automatically removes all associated visits

---

## Authentication Requirements

### Token Format
- **Type**: JWT Bearer Token
- **Header Name**: `Authorization`
- **Format**: `Bearer <token>`
- **Source**: Supabase authentication

### Required for Endpoints
- ✅ **Create Itinerary**: Authentication required
- ✅ **Update Itinerary**: Authentication required (must be owner)
- ✅ **Delete Itinerary**: Authentication required (must be owner)
- ✅ **Get My Itineraries**: Authentication required
- ❌ **List All Itineraries**: Public (no auth needed)
- ❌ **Get Itinerary by ID**: Public (no auth needed)
- ❌ **Get Itinerary Details**: Public (no auth needed)

### Token Expiration Handling
1. Monitor token expiration (check `exp` claim)
2. Refresh token before expiration (typically 1 hour)
3. Handle 401 Unauthorized responses
4. Redirect to login on authentication failure
5. Store token securely (HttpOnly cookies recommended)

---

## API Endpoints

### Base URL
```
http://localhost:8080/v1/itineraries
```

### Endpoint Summary

| Method | Endpoint | Auth | Purpose | Response Time |
|--------|----------|------|---------|---------------|
| POST | `/itineraries` | ✅ | Create new itinerary | 5-10 seconds* |
| GET | `/itineraries/:id` | ❌ | Get basic info | < 100ms |
| GET | `/itineraries/:id/details` | ❌ | Get with full place details | < 200ms |
| GET | `/itineraries/me` | ✅ | List user's itineraries | < 150ms |
| GET | `/itineraries` | ❌ | List all itineraries | < 150ms |
| PUT | `/itineraries/:id` | ✅ | Update itinerary | < 100ms |
| DELETE | `/itineraries/:id` | ✅ | Delete itinerary | < 100ms |

*Route optimization can take 5-10 seconds for 3-5 places

---

## Data Models

### 1. Create Itinerary Request

**Endpoint**: `POST /v1/itineraries`

**Required Fields**:
```
{
  "title": string (min: 2, max: 200 characters),
  "planned_date": ISO8601 datetime string,
  "start_location": {
    "latitude": decimal number (-90 to 90),
    "longitude": decimal number (-180 to 180)
  },
  "transport_mode": enum ["WALKING", "DRIVING", "CYCLING", "PUBLIC_TRANSIT"],
  "selected_places": array of place IDs (min: 2, max: 20),
  "default_duration": integer (minutes, default: 45)
}
```

**Optional Fields**:
```
{
  "description": string (max: 5000 characters),
  "visit_durations": {
    "<place_id>": integer (minutes, overrides default_duration)
  },
  "max_distance_km": decimal (filter out distant places),
  "time_window": {
    "start": "HH:MM" (e.g., "09:00"),
    "end": "HH:MM" (e.g., "17:00")
  }
}
```

**Field Validations**:
- `title`: Cannot be empty, trim whitespace
- `planned_date`: Must be current or future date
- `start_location`: Valid GPS coordinates (check bounds)
- `selected_places`: Must contain at least 2 valid place IDs
- `default_duration`: Between 15-180 minutes
- `transport_mode`: Case-sensitive, must be uppercase

### 2. Itinerary Response (Basic)

**Returned by**: GET `/itineraries/:id`, POST `/itineraries`, PUT `/itineraries/:id`

```
{
  "id": string (ULID with prefix "itin_"),
  "user_id": string (UUID),
  "title": string,
  "description": string | null,
  "planned_date": ISO8601 datetime,
  "start_location": {
    "latitude": decimal string,
    "longitude": decimal string
  },
  "transport_mode": enum,
  "total_distance_km": decimal number,
  "total_duration_minutes": integer,
  "total_visit_time_minutes": integer,
  "is_optimized": boolean,
  "status": enum ["draft", "published", "archived"],
  "created_at": ISO8601 datetime,
  "updated_at": ISO8601 datetime
}
```

**Key Metrics**:
- `total_distance_km`: Sum of distances between all consecutive places
- `total_duration_minutes`: Total travel time + visit time
- `total_visit_time_minutes`: Sum of time spent at each place (excludes travel)
- `is_optimized`: Always true (route is automatically optimized)

### 3. Itinerary Response (With Details)

**Returned by**: GET `/itineraries/:id/details`

Includes all fields from basic response plus:
```
{
  ...basic fields,
  "visits": [
    {
      "id": string (ULID with prefix "visit_"),
      "place_id": string,
      "sequence_order": integer (0-indexed),
      "planned_duration_minutes": integer,
      "distance_from_previous_km": decimal | null,
      "travel_time_from_previous_minutes": integer | null,
      "transport_mode": enum | null,
      "status": enum,
      "created_at": ISO8601 datetime,
      "updated_at": ISO8601 datetime,
      "place": {
        "id": string,
        "title": string,
        "slug": string,
        "place_type": enum,
        "short_description": string | null,
        "location": {
          "latitude": decimal string,
          "longitude": decimal string
        },
        "average_rating": decimal | null,
        "total_reviews": integer,
        "primary_image_url": string | null,
        "thumbnail_url": string | null,
        "status": enum
      }
    }
  ]
}
```

**Visit Object Details**:
- `sequence_order`: Order of visiting (0 = first place)
- `distance_from_previous_km`: Distance from previous place (null for first place)
- `travel_time_from_previous_minutes`: Travel time from previous place (null for first place)
- `place`: Full place object with all details for display

### 4. Update Itinerary Request

**Endpoint**: `PUT /v1/itineraries/:id`

**Allowed Updates** (all optional):
```
{
  "title": string,
  "description": string,
  "planned_date": ISO8601 datetime,
  "status": enum ["draft", "published", "archived"]
}
```

**Note**: Cannot update places, transport mode, or route after creation. User must create a new itinerary.

### 5. List Response (Paginated)

**Returned by**: GET `/itineraries`, GET `/itineraries/me`

```
{
  "itineraries": [
    {...basic itinerary objects}
  ],
  "pagination": {
    "total": integer,
    "limit": integer,
    "offset": integer,
    "has_more": boolean
  }
}
```

**Query Parameters**:
- `limit`: Items per page (default: 20, max: 100)
- `offset`: Skip N items (default: 0)
- `sort`: Field to sort by (e.g., "created_at", "planned_date")
- `order`: "asc" or "desc" (default: "desc")
- `status`: Filter by status

---

## Integration Workflow

### Flow 1: Creating an Itinerary (Step-by-Step)

#### Step 1: Fetch Available Places
**Purpose**: Get places that user can add to itinerary

**API Call**:
- Endpoint: `GET /v1/places`
- Query params: `?limit=50&status=published`
- Headers: None (public endpoint)

**What to Display**:
- Place cards with: title, image, location, rating
- Filter options: category, place_type, distance from user
- Search functionality
- Map view with markers

**User Actions**:
- Browse places list/grid
- Search by name or category
- Filter by distance, rating, type
- View on map
- Select places (minimum 2, maximum 20)

#### Step 2: Select Start Location
**Purpose**: Define where the journey begins

**Options**:
1. **Current Location** (recommended):
   - Request geolocation permission
   - Use browser's Geolocation API
   - Show accuracy indicator
   - Fallback to Nashik city center if denied

2. **Manual Selection**:
   - Click on map to set location
   - Search for address/landmark
   - Use saved locations (if feature exists)

3. **First Place as Start**:
   - Option to start from first selected place
   - Calculate distance from actual start

**Validation**:
- Latitude: -90 to 90
- Longitude: -180 to 180
- Should be within reasonable distance of Nashik (optional check)

#### Step 3: Configure Itinerary Details

**Form Fields**:

1. **Title** (required):
   - Input type: text
   - Placeholder: "My Nashik Temple Tour"
   - Max length: 200 characters
   - Validation: Cannot be empty or only whitespace

2. **Description** (optional):
   - Input type: textarea
   - Placeholder: "A spiritual journey through..."
   - Max length: 5000 characters
   - Auto-resize as user types

3. **Planned Date** (required):
   - Input type: date-time picker
   - Default: Tomorrow at 9:00 AM
   - Validation: Cannot be in the past
   - Show day of week and relative time ("in 2 days")

4. **Transport Mode** (required):
   - Input type: radio buttons or segmented control
   - Options:
     * 🚶 WALKING - Best for nearby places (< 5km total)
     * 🚗 DRIVING - Best for distant places or multiple stops
     * 🚴 CYCLING - Moderate distances with good roads
     * 🚌 PUBLIC_TRANSIT - Using local buses

   - Show recommended mode based on total distance
   - Display estimated time for each mode

5. **Default Visit Duration** (required):
   - Input type: slider or number input
   - Default: 45 minutes
   - Range: 15-180 minutes
   - Show in hours and minutes format

6. **Custom Visit Durations** (optional):
   - Show list of selected places
   - Allow individual duration override
   - Visual indicator for places with custom duration

**Preview Section**:
- Show selected places count
- Display approximate total distance (if possible)
- Estimated total time
- Map preview with route polyline

#### Step 4: Submit and Show Progress

**Loading State** (5-10 seconds):

Display progress indicators:
1. **Step 1**: "Validating selected places..." (1s)
2. **Step 2**: "Fetching place details..." (1-2s)
3. **Step 3**: "Calculating distances..." (2-3s)
4. **Step 4**: "Optimizing route..." (2-4s)
5. **Step 5**: "Creating itinerary..." (1s)

**UI Recommendations**:
- Show animated loader (spinning circle, progress bar)
- Display current step with descriptive text
- Disable form inputs during creation
- Provide cancel button (aborts request)
- Show estimated time remaining

**Request Configuration**:
- Method: POST
- Endpoint: `/v1/itineraries`
- Headers: `Authorization: Bearer <token>`
- Content-Type: `application/json`
- Timeout: 15 seconds (longer than usual due to optimization)

#### Step 5: Handle Response

**Success Response** (HTTP 201):

1. **Parse Response**:
   - Extract itinerary ID
   - Store in local state/cache
   - Calculate summary metrics

2. **Navigate to Detail Page**:
   - Route: `/itineraries/:id`
   - Pass itinerary data via state (avoid extra API call)
   - Show success toast/notification

3. **Display Key Information**:
   - Itinerary title
   - Total distance and time
   - Number of places
   - Planned date
   - Start location

4. **Action Buttons**:
   - View Details
   - Share Itinerary
   - Edit Details
   - Add to Calendar
   - Get Directions

**Error Response**:
- See [Error Handling](#error-handling) section

---

### Flow 2: Viewing Itinerary Details

#### Step 1: Fetch Full Details

**API Call**:
- Endpoint: `GET /v1/itineraries/:id/details`
- Headers: None (public endpoint)
- Purpose: Get itinerary with all visits and place details

**Loading State**:
- Show skeleton screen with placeholder cards
- Display loading indicators for map and route
- Estimated load time: < 200ms

#### Step 2: Display Information

**Header Section**:
- Title (editable if owner)
- Description
- Planned date with countdown ("in 3 days")
- Total metrics: distance, duration, visit time
- Status badge
- Created by username (if available)
- Share button
- Edit/Delete buttons (if owner)

**Route Overview Map**:
- Show all places as numbered markers
- Draw polyline connecting places in order
- Show start location as special marker
- Add distance labels on route segments
- Zoom to fit all places
- Interactive (click marker to highlight place)

**Places List (Ordered by Visit Sequence)**:

For each place (visit), show:

1. **Sequence Number**: 1, 2, 3...

2. **Travel Info** (except for first place):
   - Distance from previous place
   - Travel time from previous place
   - Transport mode icon
   - Direction arrow/line

3. **Place Card**:
   - Thumbnail image
   - Place name (clickable to place detail page)
   - Place type badge
   - Average rating and review count
   - Short description
   - Location (lat/lng or address)

4. **Visit Duration**:
   - Planned time at this place
   - Suggested activities (if available)
   - Best time to visit note

5. **Action Buttons**:
   - View Place Details
   - Get Directions
   - Mark as Visited (if past trip)
   - Add Review (if visited)

**Summary Section**:
- Total distance traveled
- Total time (travel + visits)
- Start time suggestion
- End time estimate
- Best season to visit
- Weather forecast (if available)
- Nearby amenities (restaurants, parking, ATMs)

#### Step 3: Interactive Features

**Map Interactions**:
- Click place marker → scroll to place card
- Hover marker → show place name tooltip
- Toggle between map and satellite view
- Show/hide route polyline
- Recenter map button
- Full-screen mode

**Timeline View** (Alternative to List):
- Vertical timeline showing visit sequence
- Time markers showing duration at each place
- Travel time between places
- Total timeline from start to end
- Current time indicator (if trip is today)

**Export Options**:
- Download as PDF
- Export to Google Calendar
- Share via link (generates shareable URL)
- Send via email
- Share on social media

---

### Flow 3: Listing User's Itineraries

#### Purpose
Show all itineraries created by logged-in user

#### API Call
**Endpoint**: `GET /v1/itineraries/me`

**Query Parameters**:
```
?limit=20          // Items per page
&offset=0          // Pagination offset
&sort=planned_date // Sort field
&order=desc        // Sort direction
&status=published  // Filter by status
```

#### Display Layout

**View Options**:
1. **Grid View**: Cards with thumbnail map
2. **List View**: Compact rows with key info
3. **Timeline View**: Sorted by planned date

**Each Itinerary Card Shows**:
- Title
- Thumbnail map of route
- Number of places
- Planned date (with relative time)
- Total distance and duration
- Status badge (draft, published)
- Last updated time
- Quick actions: View, Edit, Delete, Share

**Filtering Options**:
- Status: All, Draft, Published, Archived
- Date range: Upcoming, Past, This Month
- Transport mode: All, Walking, Driving, etc.
- Sort by: Created date, Planned date, Distance, Duration

**Empty State**:
- "You haven't created any itineraries yet"
- Call-to-action button: "Create Your First Itinerary"
- Suggestions for popular routes
- Tutorial video or guide

**Pagination**:
- Load more button (infinite scroll)
- Or traditional page numbers
- Show total count: "12 of 45 itineraries"

---

### Flow 4: Updating an Itinerary

#### What Can Be Updated
- ✅ Title
- ✅ Description
- ✅ Planned date
- ✅ Status (draft, published, archived)
- ❌ Selected places (cannot modify)
- ❌ Transport mode (cannot modify)
- ❌ Start location (cannot modify)
- ❌ Visit order (cannot modify)

**Reason**: Route optimization is complex and time-consuming. For route changes, user should create a new itinerary.

#### Update Flow

**Step 1: Enable Edit Mode**:
- Click "Edit" button
- Show editable fields inline or in modal
- Preserve original values as placeholders

**Step 2: Make Changes**:
- Update title (max 200 chars)
- Update description (max 5000 chars)
- Change planned date (date picker)
- Change status (dropdown)

**Step 3: Save Changes**:
- API Call: `PUT /v1/itineraries/:id`
- Include only changed fields in request body
- Show saving indicator
- Response time: < 100ms

**Step 4: Optimistic Update**:
- Update UI immediately (optimistic)
- Revert if API call fails
- Show success message
- Disable edit mode

**Validation**:
- Cannot set empty title
- Planned date cannot be in past (for published itineraries)
- Show inline error messages

---

### Flow 5: Deleting an Itinerary

#### Confirmation Required
**Why**: Deletion is permanent and cascades to all visits

**Confirmation Dialog**:
```
Title: "Delete Itinerary?"
Message: "This will permanently delete '${itinerary.title}' 
         and cannot be undone. All visit data will be removed."
         
Buttons: 
  - "Cancel" (default, ESC key)
  - "Delete" (destructive, requires second click or hold)
```

**Additional Checks**:
- For upcoming trips: "This trip is scheduled in 2 days. Are you sure?"
- For shared trips: "This itinerary has been shared with 3 people. They will no longer be able to access it."

#### Delete Flow

**Step 1: User Confirms**:
- Click delete button
- Show confirmation dialog
- Require explicit confirmation

**Step 2: API Call**:
- Method: DELETE
- Endpoint: `/v1/itineraries/:id`
- Headers: Authorization required
- Response: HTTP 204 (No Content)

**Step 3: Update UI**:
- Remove itinerary from list (optimistic)
- Show undo option for 5 seconds
- If undo clicked, send POST request to recreate
- Show success toast: "Itinerary deleted"

**Step 4: Handle Failure**:
- Restore itinerary to list
- Show error message
- Log error for debugging

**Cascade Effect**:
- All visits are automatically deleted
- No orphaned data
- No need for manual cleanup

---

## User Journey & UI Flows

### Complete User Journey Map

#### 1. Discovery Phase
**User Goal**: Find places to visit in Nashik

**Screens**:
1. Home/Dashboard
2. Places Listing
3. Place Detail Page
4. Favorites/Wishlist

**Actions**:
- Browse categories (Temples, Parks, Museums)
- Search places
- View on map
- Read reviews
- Check photos
- Add to favorites

#### 2. Planning Phase
**User Goal**: Create an itinerary from selected places

**Screens**:
1. Create Itinerary Form
2. Place Selection Interface
3. Map with Selected Places
4. Configuration Panel

**Actions**:
- Select 2-20 places
- Set start location
- Choose transport mode
- Set planned date
- Configure visit durations
- Preview route

**Decision Points**:
- How many places? (affects trip length)
- Walking or driving? (affects time and route)
- When to go? (affects traffic and crowds)
- How long at each place? (affects total duration)

#### 3. Review Phase
**User Goal**: Review optimized itinerary before finalizing

**Screens**:
1. Itinerary Preview
2. Route Visualization
3. Time Estimates

**Information Shown**:
- Optimized visit order (may differ from selection order)
- Distance between each place
- Total distance and time
- Estimated start/end times
- Route on map

**Actions**:
- Review and accept
- Go back to modify
- Save as draft
- Share with friends

#### 4. Execution Phase
**User Goal**: Follow itinerary during the trip

**Screens**:
1. Active Trip View
2. Navigation Interface
3. Current Place Details
4. Next Place Preview

**Features**:
- Current location tracking
- Turn-by-turn navigation
- Progress indicator (3 of 5 places visited)
- Time remaining
- Next place ETA
- Weather updates
- Quick access to phone, directions

**Actions**:
- Mark place as visited
- Skip place
- Add notes
- Take photos
- Write review
- Share on social media

#### 5. Post-Trip Phase
**User Goal**: Remember the trip and share experience

**Screens**:
1. Trip Summary
2. Photo Gallery
3. Review Form

**Actions**:
- View trip statistics
- Upload photos
- Write reviews for visited places
- Share trip report
- Save as template for future trips

---

## Error Handling

### Common Error Scenarios

#### 1. Authentication Errors (401)

**Scenario**: Token expired or invalid

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "Authentication required",
    "code": "AUTH_REQUIRED"
  }
}
```

**Frontend Handling**:
1. Detect 401 status code
2. Clear stored token
3. Redirect to login page
4. Store current route for post-login redirect
5. Show message: "Your session has expired. Please log in again."

#### 2. Validation Errors (400)

**Scenario**: Invalid input data

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "Validation failed",
    "code": "VALIDATION_ERROR",
    "details": {
      "title": "Title is required",
      "selected_places": "Must select at least 2 places",
      "planned_date": "Date cannot be in the past"
    }
  }
}
```

**Frontend Handling**:
1. Parse error details object
2. Show field-specific errors near inputs
3. Highlight invalid fields in red
4. Scroll to first error
5. Don't submit until all errors resolved

**Common Validation Errors**:
- Empty required fields
- Invalid date format
- Coordinates out of range
- Too many/few places selected
- Invalid transport mode
- Negative duration values

#### 3. Not Found Errors (404)

**Scenario**: Itinerary doesn't exist

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "Itinerary not found",
    "code": "NOT_FOUND"
  }
}
```

**Frontend Handling**:
1. Redirect to 404 error page
2. Show message: "This itinerary doesn't exist or has been deleted"
3. Offer buttons: "Go to Home", "Browse Itineraries"
4. Log error for analytics

#### 4. Permission Errors (403)

**Scenario**: User trying to modify someone else's itinerary

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "You don't have permission to modify this itinerary",
    "code": "FORBIDDEN"
  }
}
```

**Frontend Handling**:
1. Show error toast/alert
2. Hide edit/delete buttons
3. Show "View Only" badge
4. Offer "Create Copy" button instead

#### 5. Network Errors

**Scenario**: Request timeout or network failure

**Frontend Handling**:
1. Show retry button
2. Display friendly message: "Connection issue. Please check your internet."
3. Implement exponential backoff for retries
4. Cache last successful response
5. Show offline indicator

#### 6. Server Errors (500)

**Scenario**: Internal server error

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "An unexpected error occurred",
    "code": "INTERNAL_ERROR",
    "request_id": "01KC20KX3AE566GXYFM800KNKA"
  }
}
```

**Frontend Handling**:
1. Show generic error message
2. Log error with request_id
3. Provide "Report Issue" button
4. Offer to try again
5. Don't expose technical details to user

#### 7. Route Optimization Errors

**Scenario**: Cannot create feasible route

**API Response**:
```json
{
  "success": false,
  "error": {
    "message": "Cannot fit all places in the available time window",
    "code": "INFEASIBLE_ROUTE",
    "hint": "Try reducing the number of places or extending the time window",
    "details": {
      "total_duration_minutes": 480,
      "total_visit_minutes": 300,
      "places_count": 8
    }
  }
}
```

**Frontend Handling**:
1. Show specific error message with hint
2. Suggest solutions:
   - Remove some places
   - Increase visit duration limits
   - Choose faster transport mode
3. Highlight problematic places (if possible)
4. Keep form data for user to modify

### Error Display Best Practices

**Toast Notifications** (for non-critical errors):
- Position: Top-right or bottom
- Duration: 3-5 seconds
- Auto-dismiss: Yes
- Style: Colored background (red for errors)

**Modal Dialogs** (for critical errors):
- Blocks other actions
- Clear title and message
- Action buttons (Retry, Cancel, Report)
- Can't be dismissed accidentally

**Inline Errors** (for form validation):
- Next to the field with error
- Icon + text
- Appears immediately on blur or submit
- Disappears when fixed

**Page-Level Errors** (for fatal errors):
- Dedicated error page
- Friendly illustration
- Clear explanation
- Multiple action options
- Support contact information

---

## Performance Considerations

### 1. API Response Times

**Expected Latency**:
| Endpoint | Expected | Maximum | Notes |
|----------|----------|---------|-------|
| Create Itinerary | 5-7s | 10s | Route optimization is expensive |
| Get Itinerary | 50ms | 200ms | Simple database query |
| Get Details | 100ms | 300ms | Includes related data |
| Update | 50ms | 150ms | Simple update operation |
| Delete | 50ms | 150ms | Includes cascade delete |
| List | 100ms | 250ms | Depends on page size |

**Handling Slow Responses**:
1. Show loading indicators immediately
2. Display progress for create operation
3. Implement request timeout (15s for create, 5s for others)
4. Cache responses where appropriate
5. Use optimistic updates for mutations

### 2. Caching Strategy

**Client-Side Caching**:

**What to Cache**:
- ✅ List of places (cache for 10 minutes)
- ✅ Itinerary details (cache for 5 minutes)
- ✅ User's itineraries list (cache for 2 minutes)
- ❌ Create/Update responses (don't cache mutations)

**Cache Invalidation**:
- On successful create: Clear user's itineraries list cache
- On successful update: Clear specific itinerary cache
- On successful delete: Clear user's itineraries list cache
- On error: Don't update cache
- On logout: Clear all cached data

**Implementation Approaches**:
1. Browser Local Storage (simple, persistent)
2. In-memory cache (fast, cleared on page refresh)
3. Service Worker + Cache API (offline support)
4. React Query / SWR / Apollo Client (automatic cache management)

### 3. Pagination Best Practices

**Initial Load**:
- Default limit: 20 items
- Show loading skeleton for initial page
- Pre-calculate total pages if possible

**Infinite Scroll**:
- Load next page when user scrolls to 80% of current content
- Show loading indicator at bottom
- Handle edge cases (no more data, errors)
- Implement virtual scrolling for large lists

**Traditional Pagination**:
- Show page numbers (1, 2, ..., 10)
- Previous/Next buttons
- Jump to page input
- Show total: "Page 3 of 15"

**Offset Calculation**:
```
offset = (page - 1) * limit
```

Example for page 3 with limit 20:
```
offset = (3 - 1) * 20 = 40
API call: GET /itineraries?limit=20&offset=40
```

### 4. Image Optimization

**Place Images in Itinerary**:
- Use thumbnail URLs (not full-size)
- Implement lazy loading
- Show placeholder while loading
- Handle missing images gracefully
- Consider WebP format for modern browsers

**Map Rendering**:
- Use appropriate zoom level
- Tile caching for map imagery
- Optimize marker icons (small file size)
- Lazy load map library (if not needed immediately)

### 5. Network Optimization

**Request Batching**:
- Group multiple place detail requests
- Use `/itineraries/:id/details` instead of multiple API calls
- Minimize sequential requests

**Compression**:
- Enable gzip/brotli compression
- Server should send `Content-Encoding: gzip`
- Reduces payload size by 70-80%

**Request Prioritization**:
1. Critical: Authentication, initial page data
2. High: User actions (create, update, delete)
3. Medium: Secondary data (images, related content)
4. Low: Analytics, prefetching

---

## Best Practices

### 1. User Experience

**Loading States**:
- ✅ Show skeleton screens (outline of content)
- ✅ Display progress indicators with descriptive text
- ✅ Provide cancel option for long operations
- ✅ Estimate remaining time for route optimization
- ❌ Don't show generic spinners without context
- ❌ Don't freeze UI during API calls

**Success Feedback**:
- Show success toast: "Itinerary created successfully!"
- Animate transitions smoothly
- Provide next action: "View Details" button
- Add confetti animation for first itinerary creation

**Empty States**:
- Friendly illustration
- Clear explanation why it's empty
- Call-to-action button
- Sample/demo itineraries to explore

**Error Recovery**:
- Allow users to retry failed operations
- Preserve form data on errors
- Provide helpful suggestions
- Don't lose user's work

### 2. Mobile Responsiveness

**Touch Interactions**:
- Large tap targets (min 44x44 pixels)
- Swipe gestures for navigation
- Pull to refresh for lists
- Long press for context menus

**Layout Adaptations**:
- Stack cards vertically on mobile
- Collapsible sections for long content
- Bottom sheet modals instead of center modals
- Sticky headers for navigation

**Performance on Mobile**:
- Reduce image sizes
- Minimize animations
- Lazy load below-fold content
- Use system fonts for faster rendering

### 3. Accessibility

**Keyboard Navigation**:
- All interactive elements focusable
- Logical tab order
- Visible focus indicators
- Keyboard shortcuts for common actions

**Screen Reader Support**:
- Meaningful alt text for images
- ARIA labels for interactive elements
- Live regions for dynamic content
- Descriptive link text

**Visual Accessibility**:
- Sufficient color contrast (WCAG AA minimum)
- Don't rely on color alone for information
- Scalable text (respect font size settings)
- Clear visual hierarchy

### 4. Data Validation

**Client-Side Validation**:
- Validate immediately on blur
- Show inline errors
- Prevent form submission if invalid
- Use HTML5 validation attributes

**Don't Trust Client Validation Alone**:
- Server always validates
- Client validation is for UX
- Handle server validation errors
- Show server errors clearly

### 5. Security Considerations

**Token Handling**:
- Store tokens securely (HttpOnly cookies preferred)
- Don't expose tokens in URLs
- Refresh tokens before expiration
- Clear tokens on logout

**Input Sanitization**:
- Escape user input before display
- Prevent XSS attacks
- Use Content Security Policy
- Validate file uploads (if applicable)

**HTTPS Only**:
- All API calls must use HTTPS
- Mixed content warnings
- Secure cookies (Secure flag)

### 6. Internationalization (Future)

**Date Formatting**:
- Use user's locale for dates
- Support multiple time zones
- Show relative dates ("2 days ago")
- ISO 8601 for API communication

**Distance Units**:
- Support kilometers and miles
- User preference setting
- Convert on display, not in storage

**Language Support**:
- Externalize all strings
- Use i18n libraries
- Right-to-left language support
- Locale-specific formatting

### 7. Analytics & Monitoring

**Track User Actions**:
- Itinerary creation attempts
- Success/failure rates
- Average time to create
- Popular routes/places
- Drop-off points in flow

**Performance Metrics**:
- API response times
- Error rates per endpoint
- Client-side performance
- Mobile vs desktop usage

**User Feedback**:
- In-app feedback form
- Error reporting
- Feature requests
- Rating prompts

---

## Testing Recommendations

### 1. Unit Testing

**Test Cases for Itinerary Creation Form**:
- Validates required fields
- Handles valid inputs correctly
- Shows errors for invalid inputs
- Disables submit when invalid
- Enables submit when valid
- Trims whitespace from text fields
- Formats dates correctly
- Handles minimum/maximum place selections

**Test Cases for API Integration**:
- Constructs request body correctly
- Includes authentication header
- Handles success response
- Handles error responses
- Parses response data correctly
- Updates component state

### 2. Integration Testing

**Test Scenarios**:
1. **Happy Path**:
   - User selects 3 places
   - Fills in all required fields
   - Submits form
   - Sees loading state
   - Redirected to detail page
   - Data displayed correctly

2. **Error Handling**:
   - Submit with missing fields → shows validation errors
   - Token expired → redirects to login
   - Network error → shows retry option
   - Server error → shows error message

3. **Edge Cases**:
   - Minimum places (2) → should work
   - Maximum places (20) → should work
   - 21 places → should show error
   - Past date → should show validation error
   - Invalid coordinates → should reject

### 3. E2E Testing

**Full User Flows to Test**:

**Flow 1: Create and View Itinerary**:
1. Login as user
2. Navigate to create itinerary page
3. Select 4 places
4. Fill in required fields
5. Submit form
6. Wait for creation (up to 10s)
7. Verify redirect to detail page
8. Verify data displayed correctly
9. Verify map shows route
10. Verify visits are in correct order

**Flow 2: Update Itinerary**:
1. Navigate to my itineraries
2. Click on an itinerary
3. Click edit button
4. Change title
5. Change planned date
6. Save changes
7. Verify success message
8. Verify changes reflected

**Flow 3: Delete Itinerary**:
1. Navigate to my itineraries
2. Click delete on an itinerary
3. Confirm deletion
4. Verify success message
5. Verify itinerary removed from list
6. Verify 404 when accessing deleted itinerary

### 4. Performance Testing

**Metrics to Measure**:
- Time to First Contentful Paint (FCP)
- Time to Interactive (TTI)
- Largest Contentful Paint (LCP)
- Cumulative Layout Shift (CLS)
- API response times
- Bundle size

**Load Testing**:
- Test with 10, 50, 100 concurrent users
- Measure API response degradation
- Monitor server resource usage
- Test database query performance

### 5. Mobile Testing

**Devices to Test**:
- iOS (iPhone SE, iPhone 14)
- Android (Samsung Galaxy, Google Pixel)
- Tablets (iPad, Android tablet)
- Various screen sizes (320px to 1920px)

**Features to Verify**:
- Touch targets are large enough
- Forms are easy to fill
- Maps are interactive
- Scrolling is smooth
- Images load correctly
- Layouts don't break

### 6. Browser Compatibility

**Browsers to Support**:
- Chrome (latest 2 versions)
- Safari (latest 2 versions)
- Firefox (latest 2 versions)
- Edge (latest 2 versions)
- Mobile browsers (iOS Safari, Chrome Android)

**Features to Test**:
- Date picker appearance
- Geolocation API
- Local storage
- Fetch API
- ES6+ features (if not transpiled)

---

## Quick Reference

### API Endpoints Cheat Sheet

```
Authentication: Bearer Token required for ✅ marked endpoints

POST   /v1/itineraries              ✅  Create new itinerary
GET    /v1/itineraries/:id              Get itinerary basic info
GET    /v1/itineraries/:id/details      Get with full place details
GET    /v1/itineraries/me           ✅  List user's itineraries
GET    /v1/itineraries                  List all (public)
PUT    /v1/itineraries/:id          ✅  Update itinerary
DELETE /v1/itineraries/:id          ✅  Delete itinerary
```

### Common Query Parameters

```
?limit=20              Items per page (default: 20, max: 100)
?offset=0              Pagination offset
?sort=created_at       Sort field
?order=desc            Sort direction (asc/desc)
?status=published      Filter by status
```

### Transport Modes

```
WALKING          Best for: < 5km total, nearby attractions
DRIVING          Best for: Long distances, multiple stops
CYCLING          Best for: Medium distances, good roads
PUBLIC_TRANSIT   Best for: Cost-effective, eco-friendly
```

### Status Values

```
draft       Work in progress, not finalized
published   Finalized and visible to others
archived    Past trips or no longer active
```

### Common HTTP Status Codes

```
200 OK              Success (GET, PUT)
201 Created         Success (POST)
204 No Content      Success (DELETE)
400 Bad Request     Invalid input
401 Unauthorized    Authentication required
403 Forbidden       No permission
404 Not Found       Resource doesn't exist
500 Server Error    Internal server error
```

---

## Support & Resources

### Need Help?

**Technical Issues**:
- Check server logs for detailed errors
- Look for `request_id` in error responses
- Test endpoints using Postman/cURL
- Verify token is valid and not expired

**Integration Questions**:
- Review this guide thoroughly
- Check API response formats
- Verify request body structure
- Test with known working data

**Reporting Bugs**:
Include:
- Endpoint being called
- Request body (sanitize sensitive data)
- Expected vs actual response
- Error messages
- Request ID (if available)
- Screenshots or recordings

### Additional Documentation

- **Main API Documentation**: `/docs/swagger/`
- **Authentication Guide**: `/docs/AUTHENTICATION.md`
- **Places API Guide**: `/docs/PLACES_API.md`
- **General Integration**: `/docs/API_INTEGRATION_GUIDE.md`

---

## Changelog

### Version 1.0 (2025-12-09)
- ✅ Initial documentation
- ✅ Complete endpoint descriptions
- ✅ User flow diagrams
- ✅ Error handling guidelines
- ✅ Performance recommendations
- ✅ Testing guidelines
- ✅ Fixed route optimization timeout issue
- ✅ Added infinite loop prevention

### Known Issues
- ⚠️ Route optimization returns 0 visits in some cases (under investigation)
- ⚠️ Distance calculation may be inaccurate for walking routes
- ⚠️ No support for multi-day itineraries yet

---

**Document Prepared For**: Frontend Development Team  
**API Version**: v1  
**Last Updated**: December 9, 2025  
**Contact**: Backend Team

**Note**: This guide is framework/language agnostic. Adapt the concepts to your specific frontend stack (React, Vue, Angular, Flutter, Swift, Kotlin, etc.)
