# Favorites

## Favorite an Asset

```shell
curl -X POST "http://localhost:8090/assets/favorite" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "01JM9RECVAMFMY137JMWXEEW9A",
    "asset_id": "01JM9R7XTHP89ZW3GF1MB8VYHB",
    "description": "Foo Favorite"
  }'
```

> The above command returns a 202 Accepted status with an empty response body.

This endpoint asynchronously marks an asset as a favorite for a user.

This API requires valid user and asset IDs that can be fetched from the respective APIs.
### HTTP Request

`POST http://localhost:8090/assets/favorite`

### Request Body

Parameter | Type | Description
--------- | ---- | -----------
user_id | string | The ID of the user
asset_id | string | The ID of the asset to favorite
description | string | Optional description for the favorite

## List User Favorites

```shell
curl "http://localhost:8090/users/01JM9RECVAMFMY137JMWXEEW9A/favorites"
```

> The above command returns JSON structured like this:

```json
{
  "status": "success",
  "data": {
    "items": [
      {
        "id": "01JM9S0DN5FQ5ZRVZ672TGNSFG",
        "description": "Foo Favorite"
      }
    ]
  }
}
```

This endpoint retrieves a list of favorites for a specific user.

### HTTP Request

`GET http://localhost:8090/users/{user_id}/favorites`

## Update Favorite

```shell
curl -X PATCH "http://localhost:8090/users/01JM9RECVAMFMY137JMWXEEW9A/favorites/01JM9S0DN5FQ5ZRVZ672TGNSFG" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description"
  }'
```

> The above command returns JSON structured like this:

```json
{
  "status": "success",
  "data": {
    "id": "01JM9S0DN5FQ5ZRVZ672TGNSFG",
    "description": "Bar Favorite"
  }
}
```

This endpoint updates a user's favorite asset. The user updating the favorite must be the one who created it.

### HTTP Request

`PATCH http://localhost:8090/users/{user_id}/favorites/{favorite_id}`

### Request Body

Parameter | Type | Description
--------- | ---- | -----------
description | string | New description for the favorite

## Delete Favorite

```shell
curl -X DELETE "http://localhost:8090/users/01JM9RECVAMFMY137JMWXEEW9A/favorites/01JM9S0DN5FQ5ZRVZ672TGNSFG"
```

> The above command returns a 204 No Content status with an empty response body.

This endpoint removes a favorite from a user's list. The user deleting the favorite must be the one who created it.

### HTTP Request

`DELETE http://localhost:8090/users/{user_id}/favorites/{favorite_id}`
