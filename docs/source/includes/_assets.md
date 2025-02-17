# Assets

## List Assets

```shell
curl "http://localhost:8090/assets?pageSize=3&maxResults=100"
```

> The above command returns JSON structured like this:

```json
{
  "status": "success",
  "data": {
    "items": [
      {
        "id": "01JM9R7XTHP89ZW3GF1MB8VYHB",
        "type": "AUDIENCE",
        "created_at": "2025-02-17T10:46:10.513951Z",
        "updated_at": "2025-02-17T10:46:10.513951Z",
        "data": {
          "gender": "Male",
          "birth_country": "Greece",
          "age_min": 28,
          "age_max": 72,
          "social_media_hours": 6862,
          "last_month_purchases": 31
        }
      },
      {
        "id": "01JM9R7XTJ4FYVQF4N22762FNP",
        "type": "CHART",
        "created_at": "2025-02-17T10:46:10.51405Z",
        "updated_at": "2025-02-17T10:46:10.51405Z",
        "data": {
          "title": "Chart 7",
          "x_axis": "X Axis",
          "y_axis": "Y Axis",
          "data": [
            57.28889831762387,
            39.08649189369422,
            97.34424124804508,
            1.5855033280341633,
            91.09326438181805
          ]
        }
      },
      {
        "id": "01JM9R7XTJ4FYVQF4N1T4GKR05",
        "type": "INSIGHT",
        "created_at": "2025-02-17T10:46:10.514037Z",
        "updated_at": "2025-02-17T10:46:10.514037Z",
        "data": {
          "insight": "Beatae hic ipsa est explicabo et."
        }
      }
    ],
    "next_page_token": "01JM9R7XTJ4FYVQF4N1T4GKR05"
  }
}
```

This endpoint retrieves a list of test assets.

### HTTP Request

`GET http://localhost:8090/assets?pageSize=10&maxResults=100`

### Query Parameters

Parameter | Default | Description
--------- | ------- | -----------
pageSize | 10 | Number of items per page (required)
maxResults | 100 | Maximum number of results to return (required)
pageToken | - | Token for pagination (optional)

### Asset Types

The API returns three types of assets:

1. Chart Assets
2. Insight Assets
3. Audience Assets

Each asset type has its own specific data structure as shown in the response example.
